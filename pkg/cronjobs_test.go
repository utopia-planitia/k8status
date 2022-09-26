package k8status

import (
	"bytes"
	"context"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_printCronjobStatus(t *testing.T) {
	yes := true
	no := false

	PassedLastSuccessfulTime := metav1.NewTime(time.Now().Add(-720 * time.Hour))
	RecentLastSuccessfulTime := metav1.NewTime(time.Now().Add(-10 * time.Hour))

	type args struct {
		details  colorWriter
		cronjobs *batchv1.CronJobList
		verbose  bool
	}
	tests := []struct {
		name        string
		args        args
		want        int
		checkHeader bool
		wantHeader  string
		wantErr     bool
	}{
		// TODO: Add test cases.
		{
			name: "CronJobList is nil yields: error",
			args: args{
				details: colorWriter{},
			},
			want:       0,
			wantHeader: "",
			wantErr:    true,
		},
		{
			name: "CronJobs from ci/lab namespaces yielding: 0 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "ci-test",
							},
							Spec: batchv1.CronJobSpec{Suspend: &no},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "test-ci-test",
							},
							Spec: batchv1.CronJobSpec{Suspend: &no},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "test-ci",
							},
							Spec: batchv1.CronJobSpec{Suspend: &no},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "lab-test",
							},
							Spec: batchv1.CronJobSpec{Suspend: &no},
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
				},
			},
			want:        0,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},
		{
			name: "Cronjobs with nil LastsuccessfulTime yielding: 52 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "test",
							},
							Spec: batchv1.CronJobSpec{
								Suspend: &no,
							},
							Status: batchv1.CronJobStatus{
								LastSuccessfulTime: nil,
							},
						},
					},
				},
			},
			want:        52,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},
		{
			name: "Cronjobs with passed LastsuccessfulTime yielding: 53 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "test",
							},
							Spec: batchv1.CronJobSpec{
								Suspend:  &no,
								Schedule: "@hourly",
							},
							Status: batchv1.CronJobStatus{
								LastSuccessfulTime: &PassedLastSuccessfulTime,
							},
						},
					},
				},
			},
			want:        53,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},

		//succcess cases - exit code 0
		{
			name: "Suspended cronJobs yielding: 0 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
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
			},
			want:        0,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},
		{
			name: "Cronjobs with recent LastsuccessfulTime yielding: 0 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
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
			},
			want:        0,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},
		{
			name: "Cronjobs with recent LastsuccessfulTime yielding: 0 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
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
			},
			want:        0,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},
		{
			name: "Cronjobs in ci namespace + one with recent LastsuccessfulTime yielding: 0 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "ci-test",
							},
							Spec: batchv1.CronJobSpec{
								Suspend: &no,
							},
						},
						batchv1.CronJob{
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
			},
			want:        0,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},
		{
			name: "One Cronjob suspended + one with recent LastsuccessfulTime yielding: 0 exit code, no error",
			args: args{
				details: colorWriter{},
				cronjobs: &batchv1.CronJobList{
					Items: []batchv1.CronJob{
						batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "test",
							},
							Spec: batchv1.CronJobSpec{
								Suspend: &yes,
							},
						},
						batchv1.CronJob{
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
			},
			want:        0,
			checkHeader: false,
			wantHeader:  "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			header := &bytes.Buffer{}

			tt.args.details.details = &bytes.Buffer{}
			tt.args.details.noColors = true

			got, err := printCronjobStatus(ctx, header, tt.args.details, tt.args.cronjobs, tt.args.verbose)
			if (err != nil) != tt.wantErr {
				t.Errorf("printCronjobStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("printCronjobStatus() = %v, want %v", got, tt.want)
			}
			if gotHeader := header.String(); tt.checkHeader && (gotHeader != tt.wantHeader) {
				t.Errorf("printCronjobStatus() = %v, want %v", gotHeader, tt.wantHeader)
			}
			//todo: decide if we want to check details...
		})
	}
}
