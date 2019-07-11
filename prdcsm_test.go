package prdcsm_test

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/jamillosantos/macchiato"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
)

func TestPrdcsm(t *testing.T) {
	log.SetOutput(ginkgo.GinkgoWriter)
	gomega.RegisterFailHandler(ginkgo.Fail)

	description := "go-prdcsm Test Suite"
	if os.Getenv("CI") == "" {
		macchiato.RunSpecs(t, description)
	} else {
		reporterOutputDir := "./test-results/go-prdcsm"
		os.MkdirAll(reporterOutputDir, os.ModePerm)
		junitReporter := reporters.NewJUnitReporter(path.Join(reporterOutputDir, "results.xml"))
		macchiatoReporter := macchiato.NewReporter()
		ginkgo.RunSpecsWithCustomReporters(t, description, []ginkgo.Reporter{macchiatoReporter, junitReporter})
	}
}
