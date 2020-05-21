package controllers

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("shasumData", func() {
	It("should be matched", func() {
		data := map[string][]byte{
			"API_KEY":  []byte("hoge"),
			"PASSWORD": []byte("fuga"),
		}
		sum := shasumData(data)
		Expect(sum).To(Equal("b6b66b55b6b03c6ee6abc0027095d38a35937eb3e6ff2dc9f2aafa846c704e3b"))
	})
})

var _ = Describe("yamlParse", func() {
	It("yaml string should be parsed", func() {
		input := []byte("--- apikey")
		result, err := yamlParse(input)
		Expect(err).Should(BeNil())
		Expect(string(result)).To(Equal("apikey"))
	})
	It("string should not be changed", func() {
		input := []byte("apikey")
		result, err := yamlParse(input)
		Expect(err).Should(BeNil())
		Expect(string(result)).To(Equal("apikey"))
	})
})

var _ = Describe("KMSSecretReconciler", func() {
	// var c client.Client
	var stopMgr chan struct{}
	var mgrStopped *sync.WaitGroup

	const timeout = time.Second * 5

	It("should setup Manager&Reconciler", func() {
		mgr, err := manager.New(cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		//c = mgr.GetClient()

		rc := &KMSSecretReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("Captain"),
			Scheme: mgr.GetScheme(),
		}

		err = rc.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		stopMgr, mgrStopped = StartTestManager(mgr)
	})

	It("should stop Manager", func() {
		close(stopMgr)
		mgrStopped.Wait()
	})
})
