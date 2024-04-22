package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bucketv1alpha1 "github.com/amarmus/awesome-operator.git/api/v1alpha1"
	"github.com/amarmus/awesome-operator.git/internal/services"
)

const bucketFinalizerName = "awesome-operator.public.net/bucket-finalizer"

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	GCPSvc services.GCPSvc
}

//+kubebuilder:rbac:groups=awesome-operator.public.net,resources=buckets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=awesome-operator.public.net,resources=buckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=awesome-operator.public.net,resources=buckets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bucket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("bucket", req.NamespacedName)

	bucket := &bucketv1alpha1.Bucket{}
	err := r.Get(ctx, req.NamespacedName, bucket)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Bucket resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get bucket")
		return ctrl.Result{}, err
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if bucket.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.

		if !containsString(bucket.ObjectMeta.Finalizers, bucketFinalizerName) {
			bucket.ObjectMeta.Finalizers = append(bucket.ObjectMeta.Finalizers, bucketFinalizerName)
			if err := r.Update(ctx, bucket); err != nil {
				log.Error(err, "Failed to update bucket finalizers")
				return ctrl.Result{}, err
			}

			// Object updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	} else {
		// The object is being deleted
		if containsString(bucket.ObjectMeta.Finalizers, bucketFinalizerName) {
			// our finalizer is present, delete bucket
			if bucket.Spec.OnDeletePolicy == bucketv1alpha1.BucketOnDeletePolicyDestroy {
				log.Info("Deleting Storage Bucket", "Bucket.Cloud", bucket.Spec.Cloud, "Bucket.Name", bucket.Name)

				switch bucket.Spec.Cloud {
				case bucketv1alpha1.BucketCloudGcp:
					err := r.deleteGCPBucket(ctx, bucket)
					if err != nil {
						log.Error(err, "Failed to delete gcp Bucket", "Bucket.Name", bucket.Name)
						return ctrl.Result{}, err
					}
				default:
					log.Info("Bucket Cloud unknown.", "Bucket.Cloud", bucket.Spec.Cloud)
					return ctrl.Result{}, nil
				}
			}

			// remove our finalizer from the list and update it.
			bucket.ObjectMeta.Finalizers = removeString(bucket.ObjectMeta.Finalizers, bucketFinalizerName)
			if err := r.Update(context.Background(), bucket); err != nil {
				log.Error(err, "Failed to delete bucket finalizer")
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// check if the storage bucket has been created yet
	if bucket.Status.CreatedAt == "" {
		// bucket not yet created
		log.Info("Creating Bucket", "Bucket.Cloud", bucket.Spec.Cloud, "Bucket.Name", bucket.Name)

		switch bucket.Spec.Cloud {
		case bucketv1alpha1.BucketCloudGcp:
			err := r.createGCPBucket(ctx, bucket)
			if err != nil {
				log.Error(err, "Failed to create gcp Bucket", "Bucket.Name", bucket.Name)
				return ctrl.Result{}, err
			}
		default:
			log.Info("Bucket Cloud unknown.", "Bucket.Cloud", bucket.Spec.Cloud)
			return ctrl.Result{}, nil
		}

		bucket.Status.CreatedAt = time.Now().Format(time.RFC3339)
		err = r.Client.Status().Update(ctx, bucket)
		if err != nil {
			log.Error(err, "Failed to update bucket status")
			return ctrl.Result{}, err
		}

		// Status updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *BucketReconciler) createGCPBucket(ctx context.Context, bucket *bucketv1alpha1.Bucket) error {
	// create bucket
	err := r.GCPSvc.CreateBucket(ctx, bucket.Spec.FullName)
	if err != nil {
		return err
	}

	return nil
}

func (r *BucketReconciler) deleteGCPBucket(ctx context.Context, bucket *bucketv1alpha1.Bucket) error {
	// delete bucket
	err := r.GCPSvc.DeleteGCPBucket(ctx, bucket.Spec.FullName)
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bucketv1alpha1.Bucket{}).
		Complete(r)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
