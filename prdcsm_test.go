package prdcsm_test

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/jamillosantos/macchiato"
	"github.com/lab259/go-prdcsm/v2"
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

func BenchmarkPrdcsm(b *testing.B) {
	var called safecounter
	producer := prdcsm.NewChannelProducer(b.N + 1)
	pool := prdcsm.NewPool(prdcsm.PoolConfig{
		Workers:  4,
		Producer: producer,
		Consumer: func(data interface{}) {
			called.inc(data.(int))
		},
	})

	for i := 0; i < b.N; i++ {
		producer.Ch <- 1
	}
	producer.Ch <- prdcsm.EOF

	err := pool.Start()
	if err != nil {
		panic(err)
	}

	if called.count() != b.N {
		b.Errorf("%d should equal %d", called.count(), b.N)
		b.Fail()
	}
}
