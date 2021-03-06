package layerx_brain_client_test

import (
	. "github.com/emc-advanced-dev/layerx/layerx-core/layerx_brain_client"

	"fmt"
	"github.com/emc-advanced-dev/layerx/layerx-core/fakes"
	"github.com/emc-advanced-dev/layerx/layerx-core/layerx_rpi_client"
	"github.com/emc-advanced-dev/layerx/layerx-core/layerx_tpi_client"
	"github.com/emc-advanced-dev/layerx/layerx-core/lxtypes"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"github.com/mesos/mesos-go/mesosproto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func PurgeFakeServer(fakeLxUrl string) error {
	resp, _, err := lxhttpclient.Post(fakeLxUrl, "/Purge", nil, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("status code was %v", resp.StatusCode), nil)
	}
	return nil
}

var _ = Describe("LayerxBrainClient", func() {

	fakeStatus1 := fakes.FakeTaskStatus("fake_task_id_1", mesosproto.TaskState_TASK_RUNNING)

	fakeStatuses := []*mesosproto.TaskStatus{fakeStatus1}

	go fakes.NewFakeCore().Start(fakeStatuses, 12349)
	brainClient := LayerXBrainClient{
		CoreURL: "127.0.0.1:12349",
	}
	lxTpi := layerx_tpi_client.LayerXTpi{
		CoreURL: "127.0.0.1:12349",
	}
	lxRpi := layerx_rpi_client.LayerXRpi{
		CoreURL: "127.0.0.1:12349",
	}

	Describe("GetNodes", func() {
		It("returns the list of known nodes", func() {
			purgeErr := PurgeFakeServer("127.0.0.1:12349")
			Expect(purgeErr).To(BeNil())
			fakeOffer1 := fakes.FakeOffer("fake_offer_id_1", "_1")
			fakeOffer2 := fakes.FakeOffer("fake_offer_id_2", "_1")
			fakeOffer3 := fakes.FakeOffer("fake_offer_id_3", "_2")
			fakeResource1 := lxtypes.NewResourceFromMesos(fakeOffer1)
			fakeResource2 := lxtypes.NewResourceFromMesos(fakeOffer2)
			fakeResource3 := lxtypes.NewResourceFromMesos(fakeOffer3)
			err := lxRpi.SubmitResource(fakeResource1)
			Expect(err).To(BeNil())
			err = lxRpi.SubmitResource(fakeResource2)
			Expect(err).To(BeNil())
			err = lxRpi.SubmitResource(fakeResource3)
			Expect(err).To(BeNil())
			fakeNode1 := lxtypes.NewNode("_1")
			err = fakeNode1.AddResource(fakeResource1)
			Expect(err).To(BeNil())
			err = fakeNode1.AddResource(fakeResource2)
			Expect(err).To(BeNil())
			fakeNode2 := lxtypes.NewNode("_2")
			err = fakeNode2.AddResource(fakeResource3)
			Expect(err).To(BeNil())
			//the actual test
			nodes, err := brainClient.GetNodes()
			Expect(err).To(BeNil())
			Expect(nodes).To(ContainElement(fakeNode1))
			Expect(nodes).To(ContainElement(fakeNode2))
		})
	})

	Describe("GetStatusUpdates", func() {
		It("returns the list of status updates", func() {
			purgeErr := PurgeFakeServer("127.0.0.1:12349")
			Expect(purgeErr).To(BeNil())
			taskProvider := &lxtypes.TaskProvider{
				Id:     "fake_task_provider_id",
				Source: "taskprovider@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(taskProvider)
			Expect(err).To(BeNil())
			fakeTask1 := fakes.FakeLXTask("fake_task_id_1", "fake_task_name", "", "echo FAKE_COMMAND")
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask1)
			Expect(err).To(BeNil())
			statuses, err := brainClient.GetStatusUpdates()
			Expect(err).To(BeNil())
			Expect(statuses).To(ContainElement(fakeStatus1))
		})
	})

	Describe("GetPendingTasks", func() {
		It("returns the list of taks in the pending pool", func() {
			purgeErr := PurgeFakeServer("127.0.0.1:12349")
			Expect(purgeErr).To(BeNil())
			fakeLxTask := fakes.FakeLXTask("fake_task_id", "fake_task_name", "", "echo FAKE_COMMAND")
			taskProvider := &lxtypes.TaskProvider{
				Id:     "fake_task_provider_id",
				Source: "taskprovider@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(taskProvider)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeLxTask)
			Expect(err).To(BeNil())

			fakeLxTask.TaskProvider = taskProvider

			tasks, err := brainClient.GetPendingTasks()
			Expect(err).To(BeNil())
			Expect(tasks).To(ContainElement(fakeLxTask))
		})
	})

	Describe("GetStagingTasks", func() {
		It("returns the list of taks in the staging pool", func() {
			purgeErr := PurgeFakeServer("127.0.0.1:12349")
			Expect(purgeErr).To(BeNil())
			fakeTask1 := fakes.FakeLXTask("fake_task_id_1", "fake_task_name", "", "echo FAKE_COMMAND")
			fakeTask2 := fakes.FakeLXTask("fake_task_id_2", "fake_task_name", "", "echo FAKE_COMMAND")
			fakeTask3 := fakes.FakeLXTask("fake_task_id_3", "fake_task_name", "", "echo FAKE_COMMAND")
			taskProvider := &lxtypes.TaskProvider{
				Id:     "fake_task_provider_id",
				Source: "taskprovider@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(taskProvider)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask1)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask2)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask3)
			Expect(err).To(BeNil())

			fakeTask1.TaskProvider = taskProvider
			fakeTask2.TaskProvider = taskProvider
			fakeTask3.TaskProvider = taskProvider

			fakeOffer1 := fakes.FakeOffer("fake_offer_id_1", "_1")
			fakeOffer2 := fakes.FakeOffer("fake_offer_id_2", "_1")
			fakeResource1 := lxtypes.NewResourceFromMesos(fakeOffer1)
			fakeResource2 := lxtypes.NewResourceFromMesos(fakeOffer2)
			err = lxRpi.SubmitResource(fakeResource1)
			Expect(err).To(BeNil())
			err = lxRpi.SubmitResource(fakeResource2)
			Expect(err).To(BeNil())
			nodes, err := lxRpi.GetNodes()
			Expect(err).To(BeNil())
			fakeNode := nodes[0]

			err = brainClient.AssignTasks(fakeNode.Id, fakeTask1.TaskId, fakeTask2.TaskId, fakeTask3.TaskId)
			Expect(err).To(BeNil())
			tasks, err := brainClient.GetStagingTasks()
			Expect(err).To(BeNil())
			fakeTask1.NodeId = fakeNode.Id
			fakeTask2.NodeId = fakeNode.Id
			fakeTask3.NodeId = fakeNode.Id
			Expect(tasks).To(ContainElement(fakeTask1))
			Expect(tasks).To(ContainElement(fakeTask2))
			Expect(tasks).To(ContainElement(fakeTask3))
		})
	})

	Describe("AssignTasks", func() {
		It("assigns the NodeId as the SlaveId on the specified tasks", func() {
			purgeErr := PurgeFakeServer("127.0.0.1:12349")
			Expect(purgeErr).To(BeNil())
			fakeTask1 := fakes.FakeLXTask("fake_task_id_1", "fake_task_name", "", "echo FAKE_COMMAND")
			fakeTask2 := fakes.FakeLXTask("fake_task_id_2", "fake_task_name", "", "echo FAKE_COMMAND")
			fakeTask3 := fakes.FakeLXTask("fake_task_id_3", "fake_task_name", "", "echo FAKE_COMMAND")
			taskProvider := &lxtypes.TaskProvider{
				Id:     "fake_task_provider_id",
				Source: "taskprovider@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(taskProvider)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask1)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask2)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask3)
			Expect(err).To(BeNil())

			fakeTask1.TaskProvider = taskProvider
			fakeTask2.TaskProvider = taskProvider
			fakeTask3.TaskProvider = taskProvider

			fakeOffer1 := fakes.FakeOffer("fake_offer_id_1", "_1")
			fakeOffer2 := fakes.FakeOffer("fake_offer_id_2", "_1")
			fakeResource1 := lxtypes.NewResourceFromMesos(fakeOffer1)
			fakeResource2 := lxtypes.NewResourceFromMesos(fakeOffer2)
			err = lxRpi.SubmitResource(fakeResource1)
			Expect(err).To(BeNil())
			err = lxRpi.SubmitResource(fakeResource2)
			Expect(err).To(BeNil())
			nodes, err := lxRpi.GetNodes()
			Expect(err).To(BeNil())
			fakeNode := nodes[0]

			err = brainClient.AssignTasks(fakeNode.Id, fakeTask1.TaskId, fakeTask2.TaskId, fakeTask3.TaskId)
			Expect(err).To(BeNil())
			tasks, err := brainClient.GetPendingTasks()
			Expect(err).To(BeNil())
			Expect(tasks).To(BeEmpty())
		})
	})

	Describe("MigrateTasks", func() {
		It("eventually moves the specified tasks from one node to another", func() {
			purgeErr := PurgeFakeServer("127.0.0.1:12349")
			Expect(purgeErr).To(BeNil())
			fakeTask1 := fakes.FakeLXTask("fake_task_id_1", "fake_task_name", "", "echo FAKE_COMMAND")
			fakeTask2 := fakes.FakeLXTask("fake_task_id_2", "fake_task_name", "", "echo FAKE_COMMAND")
			fakeTask3 := fakes.FakeLXTask("fake_task_id_3", "fake_task_name", "", "echo FAKE_COMMAND")
			taskProvider := &lxtypes.TaskProvider{
				Id:     "fake_task_provider_id",
				Source: "taskprovider@tphost:port",
			}
			err := lxTpi.RegisterTaskProvider(taskProvider)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask1)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask2)
			Expect(err).To(BeNil())
			err = lxTpi.SubmitTask("fake_task_provider_id", fakeTask3)
			Expect(err).To(BeNil())

			fakeTask1.TaskProvider = taskProvider
			fakeTask2.TaskProvider = taskProvider
			fakeTask3.TaskProvider = taskProvider

			fakeOffer1 := fakes.FakeOffer("fake_offer_id_1", "_1")
			fakeOffer2 := fakes.FakeOffer("fake_offer_id_2", "_1")
			fakeResource1 := lxtypes.NewResourceFromMesos(fakeOffer1)
			fakeResource2 := lxtypes.NewResourceFromMesos(fakeOffer2)
			err = lxRpi.SubmitResource(fakeResource1)
			Expect(err).To(BeNil())
			err = lxRpi.SubmitResource(fakeResource2)
			Expect(err).To(BeNil())
			Expect(err).To(BeNil())
			fakeOffer3 := fakes.FakeOffer("fake_offer_id_3", "_2")
			fakeOffer4 := fakes.FakeOffer("fake_offer_id_4", "_2")
			fakeResource3 := lxtypes.NewResourceFromMesos(fakeOffer3)
			fakeResource4 := lxtypes.NewResourceFromMesos(fakeOffer4)
			err = lxRpi.SubmitResource(fakeResource3)
			Expect(err).To(BeNil())
			err = lxRpi.SubmitResource(fakeResource4)
			Expect(err).To(BeNil())
			nodes, err := lxRpi.GetNodes()
			Expect(err).To(BeNil())
			Expect(len(nodes)).To(Equal(2))
			fakeNode1 := nodes[0]
			fakeNode2 := nodes[1]

			err = brainClient.AssignTasks(fakeNode1.Id, fakeTask1.TaskId, fakeTask2.TaskId, fakeTask3.TaskId)
			Expect(err).To(BeNil())
			tasks, err := brainClient.GetPendingTasks()
			Expect(err).To(BeNil())
			Expect(tasks).To(BeEmpty())
			err = brainClient.MigrateTasks(fakeNode2.Id, fakeTask1.TaskId, fakeTask2.TaskId, fakeTask3.TaskId)
			Expect(err).To(BeNil())
		})
	})
})
