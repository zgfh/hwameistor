package E2eTest

import (
	"context"
	"time"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var _ = ginkgo.Describe("test localstorage expand volume", ginkgo.Label("periodCheck"), func() {
	f := framework.NewDefaultFramework(v1alpha1.AddToScheme)
	client := f.GetClient()
	ctx := context.TODO()
	ginkgo.It("Configure the base environment", func() {
		result := configureEnvironment(ctx)
		gomega.Expect(result).To(gomega.BeNil())
		createLdc(ctx)
	})
	ginkgo.Context("create a StorageClass", func() {
		ginkgo.It("create a sc", func() {
			//create sc
			deleteObj := apiv1.PersistentVolumeReclaimDelete
			waitForFirstConsumerObj := storagev1.VolumeBindingWaitForFirstConsumer
			examplesc := &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "local-storage-hdd-lvm",
				},
				Provisioner: "lvm.hwameistor.io",
				Parameters: map[string]string{
					"replicaNumber":             "1",
					"poolClass":                 "HDD",
					"poolType":                  "REGULAR",
					"volumeKind":                "LVM",
					"striped":                   "true",
					"csi.storage.k8s.io/fstype": "xfs",
				},
				ReclaimPolicy:        &deleteObj,
				AllowVolumeExpansion: boolPter(true),
				VolumeBindingMode:    &waitForFirstConsumerObj,
			}
			err := client.Create(ctx, examplesc)
			if err != nil {
				logrus.Printf("Create SC failed ：%+v ", err)
				f.ExpectNoError(err)
			}
		})
	})
	ginkgo.Context("create a PVC", func() {
		ginkgo.It("create a PVC", func() {
			//create PVC
			storageClassName := "local-storage-hdd-lvm"
			examplePvc := &apiv1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-lvm",
					Namespace: "default",
				},
				Spec: apiv1.PersistentVolumeClaimSpec{
					AccessModes:      []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
					StorageClassName: &storageClassName,
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			}
			err := client.Create(ctx, examplePvc)
			if err != nil {
				logrus.Printf("Create PVC failed ：%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})

	})
	ginkgo.Context("create a deployment", func() {

		ginkgo.It("create deployment", func() {
			//create deployment
			exampleDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DeploymentName,
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "demo",
						},
					},
					Template: apiv1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "demo",
							},
						},
						Spec: apiv1.PodSpec{
							SchedulerName: "hwameistor-scheduler",
							Containers: []apiv1.Container{
								{
									Name:  "web",
									Image: "daocloud.io/daocloud/dao-2048:latest",
									Ports: []apiv1.ContainerPort{
										{
											Name:          "http",
											Protocol:      apiv1.ProtocolTCP,
											ContainerPort: 80,
										},
									},
									VolumeMounts: []apiv1.VolumeMount{
										{
											Name:      "2048-volume-lvm",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []apiv1.Volume{
								{
									Name: "2048-volume-lvm",
									VolumeSource: apiv1.VolumeSource{
										PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
											ClaimName: "pvc-lvm",
										},
									},
								},
							},
						},
					},
				},
			}
			err := client.Create(ctx, exampleDeployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("PVC STATUS should be Bound", func() {
			pvc := &apiv1.PersistentVolumeClaim{}
			pvcKey := k8sclient.ObjectKey{
				Name:      "pvc-lvm",
				Namespace: "default",
			}
			err := client.Get(ctx, pvcKey, pvc)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			logrus.Infof("Waiting for the PVC to be bound")
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				if err = client.Get(ctx, pvcKey, pvc); pvc.Status.Phase != apiv1.ClaimBound {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("PVC binding timeout")
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("deploy STATUS should be AVAILABLE", func() {
			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      DeploymentName,
				Namespace: "default",
			}
			err := client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Error(err)
				f.ExpectNoError(err)
			}
			logrus.Infof("waiting for the deployment to be ready ")
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				if err = client.Get(ctx, deployKey, deployment); deployment.Status.AvailableReplicas != int32(1) {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("deployment ready timeout")
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())

		})

	})
	ginkgo.Context("Test the volume", func() {
		ginkgo.It("write test data", func() {

			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      DeploymentName,
				Namespace: "default",
			}
			err = client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			apps, err := labels.NewRequirement("app", selection.In, []string{"demo"})
			selector := labels.NewSelector()
			selector = selector.Add(*apps)
			listOption := k8sclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &v1.PodList{}
			err = client.List(ctx, podlist, &listOption)

			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			containers := deployment.Spec.Template.Spec.Containers
			for _, pod := range podlist.Items {
				for _, container := range containers {
					_, _, err := ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && echo it-is-a-test >test", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					output, _, err := ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && cat test", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					gomega.Expect(output).To(gomega.Equal("it-is-a-test"))
				}
			}
		})
	})
	ginkgo.Context("Expand volume", func() {
		ginkgo.It("Modify spec-resources-requests-storage", func() {
			pvc := &apiv1.PersistentVolumeClaim{}
			pvcKey := k8sclient.ObjectKey{
				Name:      "pvc-lvm",
				Namespace: "default",
			}
			err := client.Get(ctx, pvcKey, pvc)
			if err != nil {
				logrus.Printf("Failed to find pvc ：%+v ", err)
				f.ExpectNoError(err)
			}
			storageMap := make(map[apiv1.ResourceName]resource.Quantity)
			storageMap[apiv1.ResourceStorage] = resource.MustParse("2Gi")
			pvc.Spec.Resources.Requests = storageMap
			err = client.Update(ctx, pvc)
			pvc = &apiv1.PersistentVolumeClaim{}
			pvcKey = k8sclient.ObjectKey{
				Name:      "pvc-lvm",
				Namespace: "default",
			}
			err = client.Get(ctx, pvcKey, pvc)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
			logrus.Infof(pvc.Status.Capacity.Storage().String())
			logrus.Infof("Waiting for the PVC to be bound")
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				if err = client.Get(ctx, pvcKey, pvc); pvc.Status.Capacity.Storage().String() != "2Gi" {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("PVC binding timeout")
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())

		})
		ginkgo.It("check test file", func() {
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      DeploymentName,
				Namespace: "default",
			}
			err = client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			apps, err := labels.NewRequirement("app", selection.In, []string{"demo"})
			selector := labels.NewSelector()
			selector = selector.Add(*apps)
			listOption := k8sclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &v1.PodList{}
			err = client.List(ctx, podlist, &listOption)

			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			containers := deployment.Spec.Template.Spec.Containers
			for _, pod := range podlist.Items {
				for _, container := range containers {
					output, _, err := ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && cat test", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					gomega.Expect(output).To(gomega.Equal("it-is-a-test"))
				}
			}
		})
	})
	ginkgo.Context("Clean up the environment", func() {
		ginkgo.It("Delete test Deployment", func() {
			//delete deploy
			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      DeploymentName,
				Namespace: "default",
			}
			err := client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Error(err)
				f.ExpectNoError(err)
			}
			logrus.Infof("deleting test Deployment ")

			err = client.Delete(ctx, deployment)
			if err != nil {
				logrus.Error(err)
				f.ExpectNoError(err)
			}
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				if err := client.Get(ctx, deployKey, deployment); !k8serror.IsNotFound(err) {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())

		})
		ginkgo.It("delete all pvc ", func() {
			err := deleteAllPVC(ctx)
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("delete all sc", func() {
			err := deleteAllSC(ctx)
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("delete helm", func() {
			uninstallHelm()
		})
	})

})
