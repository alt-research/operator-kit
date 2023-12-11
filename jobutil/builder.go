package jobutil

import (
	"bytes"
	"context"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/alt-research/operator-kit/envs"
	"github.com/alt-research/operator-kit/k8s"
	"github.com/alt-research/operator-kit/maputil"
	"github.com/alt-research/operator-kit/must"
	"github.com/alt-research/operator-kit/ptr"
	"github.com/alt-research/operator-kit/s3util"
	"github.com/alt-research/operator-kit/targz"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/env"
)

var DEFAULT_K8S_TOOL_IMAGE = env.GetString("K8S_TOOL_IMAGE", "alpine/k8s:1.28.4")

const (
	s3KeyPrefix = "jobutil"
)

type JobBuilder struct {
	clientset *kubernetes.Clientset

	// Job is the Job object
	Job                       *batchv1.Job
	ScriptCM                  *corev1.ConfigMap
	ScriptSourceConfigMapName string

	BucketName    string
	BucketManager *s3util.BucketManager
	ObjectKey     string
	K8sToolImage  string

	Name      string
	Namespace string
	Image     string
	Script    string
	WorkDir   string
	DataDirs  []string
	Env       []corev1.EnvVar
	LocalDir  string

	MaxRetries     *int32
	NewDataOnRetry bool

	NodeSelector   map[string]string
	Resources      corev1.ResourceRequirements
	ServiceAccount string
}

func BuilderSimple(name, image, workdir, script, localDir string, dataDirs []string, bucket string, env []corev1.EnvVar) *JobBuilder {
	return &JobBuilder{
		Name:       name,
		Image:      image,
		WorkDir:    workdir,
		Env:        env,
		DataDirs:   dataDirs,
		Script:     script,
		BucketName: bucket,
	}
}

func (j *JobBuilder) Build(ctx context.Context) (err error) {
	j.initDefaults()
	return j.build(ctx)
}

func (j *JobBuilder) initClient() (err error) {
	if j.clientset == nil {
		j.clientset, err = k8s.GetClient()
		if err != nil {
			return
		}
	}
	if j.BucketManager == nil {
		if j.BucketName == "" {
			return errors.New("bucket name is required")
		}
		j.BucketManager, err = s3util.NewManager(j.BucketName, "", 1)
		if err != nil {
			return
		}
	}
	return
}

func (j *JobBuilder) initDefaults() {
	j.K8sToolImage = must.Default(j.K8sToolImage, DEFAULT_K8S_TOOL_IMAGE)
	j.Namespace = must.Default(j.Namespace, k8s.NAMESPACE)
	j.ObjectKey = s3KeyPrefix + "/" + j.Namespace + "/" + j.Name + ".tar.gz"
	if j.MaxRetries == nil {
		j.MaxRetries = ptr.Of(int32(10))
	}
}

func (j *JobBuilder) build(ctx context.Context) (err error) {
	_ = j.Get(ctx)

	err = j.initClient()
	if err != nil {
		return
	}
	if j.ServiceAccount == "" {
		j.ServiceAccount, err = k8s.GetSelfServiceAccount(ctx, "")
		if err != nil {
			return
		}
	}
	sa, err := j.getServiceAccount(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to get service account %s", j.ServiceAccount)
	}

	downupEnvVars := []corev1.EnvVar{
		{Name: "BUCKET", Value: j.BucketManager.Bucket},
		{Name: "OBJECT_KEY", Value: j.ObjectKey},
		{Name: "NEW_DATA_ON_RETRY", Value: strconv.FormatBool(j.NewDataOnRetry)},
	}
	// support env provided aws credentials
	if _, ok := sa.Annotations["eks.amazonaws.com/role-arn"]; !ok {
		for k, v := range envs.SliceToMap(os.Environ()) {
			if strings.HasPrefix(k, "AWS_") && v != "" {
				downupEnvVars = append(downupEnvVars, corev1.EnvVar{Name: k, Value: v})
			}
		}
	}

	if j.ScriptCM == nil {
		j.ScriptCM = &corev1.ConfigMap{}
	}
	j.ScriptCM.Namespace = j.Namespace
	j.ScriptCM.Name = j.Name
	if j.ScriptCM.Data == nil {
		j.ScriptCM.Data = make(map[string]string)
	}
	j.ScriptCM.Data["starter.sh"] = j.Script
	j.ScriptCM.Data["entrypoint.sh"] = string(must.Two(assets.ReadFile("scripts/entrypoint.sh")))
	j.ScriptCM.Data["download-data.sh"] = string(must.Two(assets.ReadFile("scripts/download-data.sh")))
	j.ScriptCM.Data["upload-data.sh"] = string(must.Two(assets.ReadFile("scripts/upload-data.sh")))

	if j.ScriptSourceConfigMapName != "" {
		cm, err := j.clientset.CoreV1().ConfigMaps(j.Namespace).Get(ctx, j.ScriptSourceConfigMapName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to get configmap %s", j.ScriptSourceConfigMapName)
		}
		maputil.MergeOverwrite(&j.ScriptCM.Data, cm.Data)
	}

	if j.Job == nil {
		j.Job = &batchv1.Job{}
	}
	j.Job.Namespace = j.Namespace
	j.Job.Name = j.Name
	j.Job.Spec.BackoffLimit = j.MaxRetries
	j.Job.Spec.Template.Spec.ServiceAccountName = j.ServiceAccount
	j.Job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	j.Job.Spec.Template.Spec.NodeSelector = j.NodeSelector

	if len(j.Job.Spec.Template.Spec.Volumes) == 0 {
		j.Job.Spec.Template.Spec.Volumes = []corev1.Volume{
			{Name: "data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "marker", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "scripts", VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: j.ScriptCM.Name},
					DefaultMode:          ptr.Of(int32(0o644)),
				},
			}},
		}
	} else {
		for i := 0; i < len(j.Job.Spec.Template.Spec.Volumes); i++ {
			v := &j.Job.Spec.Template.Spec.Volumes[i]
			switch v.Name {
			case "data":
				v.VolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
			case "marker":
				v.VolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
			case "scripts":
				v.VolumeSource.ConfigMap = &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: j.ScriptCM.Name},
					DefaultMode:          ptr.Of(int32(0o644)),
				}
			}
		}
	}

	// init dir
	downloadDataDir := corev1.Container{
		Name:    "downloaddata",
		Image:   j.K8sToolImage,
		Env:     downupEnvVars,
		Command: []string{"bash", "/scripts/download-data.sh"},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data-dir"},
			{Name: "marker", MountPath: "/tmp/marker"},
			{Name: "scripts", MountPath: "/scripts"},
		},
	}
	if len(j.Job.Spec.Template.Spec.InitContainers) == 0 {
		j.Job.Spec.Template.Spec.InitContainers = []corev1.Container{
			downloadDataDir,
		}
	} else {
		for i := 0; i < len(j.Job.Spec.Template.Spec.InitContainers); i++ {
			c := &j.Job.Spec.Template.Spec.InitContainers[i]
			if c.Name != "downloaddata" {
				continue
			}
			c.Image = downloadDataDir.Image
			c.Env = downloadDataDir.Env
			c.Command = downloadDataDir.Command
			c.VolumeMounts = downloadDataDir.VolumeMounts
		}
	}

	uploadDataDir := corev1.Container{
		Name:    "uploaddata",
		Image:   j.K8sToolImage,
		Env:     downupEnvVars,
		Command: []string{"bash", "/scripts/upload-data.sh"},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data-dir"},
			{Name: "marker", MountPath: "/tmp/marker"},
			{Name: "scripts", MountPath: "/scripts"},
		},
	}
	workload := corev1.Container{
		Name:            "workload",
		Image:           j.Image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env:             j.Env,
		WorkingDir:      j.WorkDir,
		Resources:       j.Resources,
		VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data-dir"},
			{Name: "marker", MountPath: "/tmp/marker"},
			{Name: "scripts", MountPath: "/scripts"},
		},
		Command: []string{"bash", "/scripts/entrypoint.sh"},
	}
	for _, dir := range j.DataDirs {
		workload.VolumeMounts = append(workload.VolumeMounts, corev1.VolumeMount{
			Name: "data", SubPath: strings.TrimLeft(dir, "/"), MountPath: dir,
		})
	}

	if len(j.Job.Spec.Template.Spec.Containers) == 0 {
		j.Job.Spec.Template.Spec.Containers = []corev1.Container{
			workload,
			uploadDataDir,
		}
	} else {
		for i := 0; i < len(j.Job.Spec.Template.Spec.Containers); i++ {
			c := &j.Job.Spec.Template.Spec.Containers[i]
			switch c.Name {
			case "workload":
				c.Image = workload.Image
				c.ImagePullPolicy = workload.ImagePullPolicy
				c.Env = workload.Env
				c.WorkingDir = workload.WorkingDir
				c.Resources = workload.Resources
				c.VolumeMounts = workload.VolumeMounts
				c.Command = workload.Command
			case "uploaddata":
				c.Image = uploadDataDir.Image
				c.Env = uploadDataDir.Env
				c.Command = uploadDataDir.Command
				c.VolumeMounts = uploadDataDir.VolumeMounts
			}
		}
	}
	return
}

func (j *JobBuilder) getServiceAccount(ctx context.Context) (*corev1.ServiceAccount, error) {
	err := j.initClient()
	if err != nil {
		return nil, err
	}
	return j.clientset.CoreV1().ServiceAccounts(j.Namespace).Get(ctx, j.ServiceAccount, metav1.GetOptions{})
}

func (j *JobBuilder) Get(ctx context.Context) (err error) {
	j.initDefaults()
	err = j.initClient()
	if err != nil {
		return
	}
	job, _ := j.clientset.BatchV1().Jobs(j.Namespace).Get(ctx, j.Name, metav1.GetOptions{})
	// if err != nil && !apierrors.IsNotFound(err) && !strings.Contains(err.Error(), "not found") {
	// 	return
	// }
	j.Job = job
	cm, _ := j.clientset.CoreV1().ConfigMaps(j.Namespace).Get(ctx, j.Name, metav1.GetOptions{})
	// if err != nil && !apierrors.IsNotFound(err) && !strings.Contains(err.Error(), "not found") {
	// 	return
	// }
	j.ScriptCM = cm
	return
}

func (j *JobBuilder) Destroy(ctx context.Context, delJob bool) (err error) {
	err = j.initClient()
	if err != nil {
		return
	}
	delOP := metav1.DeleteOptions{PropagationPolicy: ptr.Of(metav1.DeletePropagationForeground)}
	if delJob {
		err = j.clientset.BatchV1().Jobs(j.Namespace).Delete(ctx, j.Name, delOP)
	} else {
		err = j.clientset.CoreV1().Pods(j.Namespace).DeleteCollection(ctx, delOP, metav1.ListOptions{
			LabelSelector: batchv1.JobNameLabel + "=" + j.Name,
		})
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return
	}
	err = j.clientset.CoreV1().ConfigMaps(j.Namespace).Delete(ctx, j.Name, delOP)
	if err != nil && !apierrors.IsNotFound(err) {
		return
	}
	return nil
}

func (j *JobBuilder) DeleteData(ctx context.Context) (err error) {
	err = j.initClient()
	if err != nil {
		return
	}
	return j.BucketManager.DeleteSingle(ctx, j.ObjectKey)
}

func (j *JobBuilder) CreateOrUpdate(ctx context.Context) (err error) {
	err = j.initClient()
	if err != nil {
		return
	}

	_, err = j.clientset.CoreV1().ConfigMaps(j.Namespace).Create(ctx, j.ScriptCM, metav1.CreateOptions{})
	if err != nil {
		_, err = j.clientset.CoreV1().ConfigMaps(j.Namespace).Update(ctx, j.ScriptCM, metav1.UpdateOptions{})
	}
	if err != nil {
		return
	}

	_, err = j.clientset.BatchV1().Jobs(j.Namespace).Create(ctx, j.Job, metav1.CreateOptions{})
	if err != nil {
		_, err = j.clientset.BatchV1().Jobs(j.Namespace).Update(ctx, j.Job, metav1.UpdateOptions{})
	}
	if err != nil {
		return
	}

	err = j.Get(ctx)
	if err != nil {
		return
	}
	// set configmap's owner to job
	j.ScriptCM.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: "batch/v1",
		Kind:       "Job",
		Name:       j.Name,
		UID:        j.Job.UID,
	}}
	_, err = j.clientset.CoreV1().ConfigMaps(j.Namespace).Update(ctx, j.ScriptCM, metav1.UpdateOptions{})
	return
}

func (j *JobBuilder) UploadData(ctx context.Context, src ...string) (err error) {
	if len(src) == 0 {
		src = []string{j.LocalDir}
	}
	// if src not exits, skip
	if _, err = os.Stat(src[0]); os.IsNotExist(err) {
		return
	}
	err = j.initClient()
	if err != nil {
		return
	}
	tempfile := must.Two(os.CreateTemp("", "jobutil-*.tar.gz"))
	_ = tempfile.Close()
	defer func() {
		_ = os.Remove(tempfile.Name())
	}()
	err = targz.Compress(src[0], tempfile.Name())
	if err != nil {
		return
	}
	_, err = j.BucketManager.Upload(ctx, tempfile.Name(), j.ObjectKey, nil, nil)
	return
}

func (j *JobBuilder) ObjectS3URL() string {
	return j.BucketManager.ObjectS3URL(j.ObjectKey)
}

func (j *JobBuilder) ObjectInfo(ctx context.Context) (info s3util.HeadObjectOutput, err error) {
	err = j.initClient()
	if err != nil {
		return
	}
	return j.BucketManager.HeadObject(ctx, j.ObjectKey)
}

func (j *JobBuilder) DownloadData(ctx context.Context, dest ...string) (err error) {
	err = j.initClient()
	if err != nil {
		return
	}
	if len(dest) == 0 {
		dest = []string{j.LocalDir}
	}
	tempfile := must.Two(os.CreateTemp("", "jobutil-*.tar.gz"))
	defer func() {
		_ = os.Remove(tempfile.Name())
	}()
	_, err = j.BucketManager.DownloadWriter(ctx, j.ObjectKey, tempfile)
	if err != nil {
		return
	}
	_ = tempfile.Close()
	// untar tar.gz to dest dir
	err = os.MkdirAll(dest[0], 0o755)
	if err != nil {
		return
	}
	return targz.Extract(tempfile.Name(), dest[0])
}

func (j *JobBuilder) GetLogs(ctx context.Context) (string, error) {
	err := j.initClient()
	if err != nil {
		return "", err
	}
	pods, err := j.clientset.CoreV1().Pods(j.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: batchv1.JobNameLabel + "=" + j.Name,
	})
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", nil
	}
	req := j.clientset.CoreV1().Pods(j.Namespace).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{
		Container:  "workload",
		TailLines:  ptr.Of(int64(1024)),
		LimitBytes: ptr.Of(int64(1024 * 1024)),
	})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "error in opening stream", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "error in copy information from podLogs to buf", err
	}
	return buf.String(), nil
}
