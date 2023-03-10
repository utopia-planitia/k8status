package k8status

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_cronjobsStatus_ExitCode(t *testing.T) {
	yes := true
	no := false

	PassedLastSuccessfulTime := metav1.NewTime(time.Now().Add(-720 * time.Hour))
	RecentLastSuccessfulTime := metav1.NewTime(time.Now().Add(-10 * time.Hour))

	tests := []struct {
		name     string
		cronjobs []batchv1.CronJob
		want     int
	}{
		{
			name: "CronJobs from ci/lab namespaces yielding: 0 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ci-test",
					},
					Spec:   batchv1.CronJobSpec{Suspend: &no},
					Status: batchv1.CronJobStatus{LastSuccessfulTime: nil},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-ci-test",
					},
					Spec:   batchv1.CronJobSpec{Suspend: &no},
					Status: batchv1.CronJobStatus{LastSuccessfulTime: nil},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-ci",
					},
					Spec:   batchv1.CronJobSpec{Suspend: &no},
					Status: batchv1.CronJobStatus{LastSuccessfulTime: nil},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "lab-test",
					},
					Spec:   batchv1.CronJobSpec{Suspend: &no},
					Status: batchv1.CronJobStatus{LastSuccessfulTime: nil},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-lab-test",
					},
					Spec: batchv1.CronJobSpec{Suspend: &no},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-lab",
					},
					Spec: batchv1.CronJobSpec{Suspend: &no},
				},
			},
			want: 0,
		},
		{
			name: "Cronjobs with no LastsuccessfulTime and never scheduled yielding: 0 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend: &no,
					},
					Status: batchv1.CronJobStatus{
						LastSuccessfulTime: nil,
						LastScheduleTime:   nil,
					},
				},
			},
			want: 0,
		},
		{
			name: "Cronjobs with no LastsuccessfulTime yielding: 52 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend: &no,
					},
					Status: batchv1.CronJobStatus{
						LastSuccessfulTime: nil,
						LastScheduleTime:   &metav1.Time{},
					},
				},
			},
			want: 52,
		},
		{
			name: "Cronjobs with passed LastsuccessfulTime yielding: 53 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend:  &no,
						Schedule: "@hourly",
					},
					Status: batchv1.CronJobStatus{
						LastSuccessfulTime: &PassedLastSuccessfulTime,
						LastScheduleTime:   &metav1.Time{},
					},
				},
			},
			want: 52,
		},

		//succcess cases - exit code 0
		{
			name: "Suspended cronJobs yielding: 0 exit code, no error",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{Suspend: &yes},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ci-test",
					},
					Spec: batchv1.CronJobSpec{Suspend: &no},
				},
			},
		},
		{
			name: "Cronjobs with recent LastsuccessfulTime yielding: 0 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend:  &no,
						Schedule: "@hourly",
					},
					Status: batchv1.CronJobStatus{
						LastSuccessfulTime: &RecentLastSuccessfulTime,
					},
				},
			},
		},
		{
			name: "Cronjobs with recent LastsuccessfulTime yielding: 0 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend:  &no,
						Schedule: "@hourly",
					},
					Status: batchv1.CronJobStatus{
						LastSuccessfulTime: &RecentLastSuccessfulTime,
					},
				},
			},
		},
		{
			name: "Cronjobs in ci namespace + one with recent LastsuccessfulTime yielding: 0 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ci-test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend: &no,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend:  &no,
						Schedule: "@hourly",
					},
					Status: batchv1.CronJobStatus{
						LastSuccessfulTime: &RecentLastSuccessfulTime,
					},
				},
			},
		},
		{
			name: "One Cronjob suspended + one with recent LastsuccessfulTime yielding: 0 exit code",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend: &yes,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
					},
					Spec: batchv1.CronJobSpec{
						Suspend:  &no,
						Schedule: "@hourly",
					},
					Status: batchv1.CronJobStatus{
						LastSuccessfulTime: &RecentLastSuccessfulTime,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := cronjobsStatus{
				cronjobs: []batchv1.CronJob{},
			}
			status.add(tt.cronjobs)

			got := status.ExitCode()
			if got != tt.want {
				t.Errorf("printCronjobStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
