package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPoseidon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Poseidon Suite")
}
