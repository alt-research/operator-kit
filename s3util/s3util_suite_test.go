package s3util_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestS3util(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3util Suite")
}

var _ = BeforeSuite(func() {
	By("Test Setup")
})
var _ = AfterSuite(func() {
	By("Test Teardown")
})
