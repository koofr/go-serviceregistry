package serviceregistry_test

import (
	"git.koofr.lan/go-netutils.git"
	. "git.koofr.lan/go-serviceregistry.git"
	"git.koofr.lan/go-zkutils.git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zk "launchpad.net/gozk"
	"strconv"
)

var _ = Describe("Zkregistry", func() {
	var s *zkutils.TestServer
	var r *ZkRegistry
	var z *zk.Conn

	BeforeEach(func() {
		port, err := netutils.UnusedPort()
		Expect(err).NotTo(HaveOccurred())

		s, err = zkutils.NewTestServer(port)
		Expect(err).NotTo(HaveOccurred())

		if err != nil {
			return
		}

		zz, session, err := zk.Dial("localhost:"+strconv.Itoa(port), 5e9)
		Expect(err).NotTo(HaveOccurred())

		Expect((<-session).State).To(Equal(zk.STATE_CONNECTED))

		z = zz

		rr, err := NewZkRegistry("localhost:" + strconv.Itoa(port))
		Expect(err).NotTo(HaveOccurred())

		r = rr
	})

	AfterEach(func() {
		z.Close()

		if r != nil {
			r.Close()
		}

		s.Stop()
	})

	It("should register service", func() {
		stat, err := z.Exists("/services/myservice/proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(stat).To(BeNil())

		err = r.Register("myservice", "proto", "localhost:1234")
		Expect(err).NotTo(HaveOccurred())

		stat, err = z.Exists("/services/myservice/proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(stat).NotTo(BeNil())

		children, _, err := z.Children("/services/myservice/proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(children).To(HaveLen(1))

		data, _, err := z.Get("/services/myservice/proto/" + children[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal("localhost:1234"))

		err = r.Register("myservice", "proto", "localhost:1235")
		Expect(err).NotTo(HaveOccurred())

		children, _, err = z.Children("/services/myservice/proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(children).To(HaveLen(2))
	})

	It("should remove service when registry closes", func() {
		err := r.Register("myservice", "proto", "localhost:1234")
		Expect(err).NotTo(HaveOccurred())

		children, _, err := z.Children("/services/myservice/proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(children).To(HaveLen(1))

		r.Close()

		children, _, err = z.Children("/services/myservice/proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(children).To(HaveLen(0))
	})

	It("should get service", func() {
		servers, err := r.Get("myservice", "proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(servers).To(Equal([]string{}))

		err = r.Register("myservice", "proto", "localhost:1234")
		Expect(err).NotTo(HaveOccurred())

		servers, err = r.Get("myservice", "proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(servers).To(Equal([]string{"localhost:1234"}))

		err = r.Register("myservice", "proto", "localhost:1235")
		Expect(err).NotTo(HaveOccurred())

		servers, err = r.Get("myservice", "proto")
		Expect(err).NotTo(HaveOccurred())
		Expect(servers).To(Equal([]string{"localhost:1234", "localhost:1235"}))
	})
})
