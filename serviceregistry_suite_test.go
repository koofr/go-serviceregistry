package serviceregistry_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServiceregistry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceRegistry Suite")
}
