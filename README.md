# Operator Kit (opkit)

A set of utilities to help writting operators for Kubernetes.

## Operator Quick Start

### 1. Install Kubebuilder

https://book.kubebuilder.io/quick-start.html#installation

```bash
# download kubebuilder and install locally.
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && mv kubebuilder /usr/local/bin/
```

### 2. Create a new project

```bash
kubebuilder init --domain altlayer.io --license mit --owner "AltResearch"
```

### 3. Create an API

```bash
kubebuilder create api --group node --version v1alpha1 --kind Verifier
```

### 4. Download Operator Kit

```bash
go get -u github.com/altresearch/operator-kit
```

### 4. Implement the API

1. every API should have ConditionPhase in its `status` field

   ```go
    import . "github.com/altresearch/operator-kit/commonspec"

    type VerifierStatus struct {
        //+optional
        ConditionPhase `json:",inline"`
    }
   ```

2. Use `specutil.ConditionManager` to manage your reconcilation
3. Use `specutil.NewControllerManagedBy` to manage your event watching

## TODO:

- [ ] Check and remove sensitive data and open this project
- [ ] Add more examples and docs
- [ ] Move specutil and commonspec to project root namespace
