package extractimageurls

import (
	"reflect"
	"strings"
	"testing"
)

func TestFromTerraform(t *testing.T) {
	tc := []struct {
		Name           string
		FileContent    string
		ExpectedImages []string
		WantErr        bool
	}{
		{
			Name:           "cloud run service, pass",
			ExpectedImages: []string{"gcr.io/google-samples/hello-app:1.0"},
			WantErr:        false,
			FileContent: `resource "google_cloud_run_service" "run_service" {
  name = "app"
  location = "us-central1"

  template {
    spec {
      containers {
        image =  "gcr.io/google-samples/hello-app:1.0"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}`,
		},
		{
			Name:           "complex terraform, pass",
			ExpectedImages: []string{"us-docker.pkg.dev/cloudrun/container/hello:latest"},
			WantErr:        false,
			FileContent: `resource "google_cloud_run_service" "default" {
  name     = "cloudrun-srv"
  location = "us-central1"

  template {
    spec {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello:latest"
      }
    }
  }
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  location    = google_cloud_run_service.default.location
  project     = google_cloud_run_service.default.project
  service     = google_cloud_run_service.default.name

  policy_data = data.google_iam_policy.noauth.policy_data
}`,
		},
		{
			Name:           "multiple cloud runs, pass",
			ExpectedImages: []string{"gcr.io/google-samples/hello-app:1.0", "us-docker.pkg.dev/cloudrun/container/hello:latest"},
			WantErr:        false,
			FileContent: `resource "google_cloud_run_service" "run_service" {
  name = "app"
  location = "us-central1"

  template {
    spec {
      containers {
        image =  "gcr.io/google-samples/hello-app:1.0"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

resource "google_cloud_run_service" "run_service" {
  name = "app"
  location = "us-central1"

  template {
    spec {
      containers {
        image =  "us-docker.pkg.dev/cloudrun/container/hello:latest"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}`,
		},
		{
			Name:           "cloud run service with path in comment, pass",
			ExpectedImages: []string{"gcr.io/google-samples/hello-app:1.0"},
			WantErr:        false,
			FileContent: `# development/tools/cmd/dnscollector
resource "google_cloud_run_service" "run_service" {
  name = "app"
  location = "us-central1"

  template {
    spec {
      containers {
        image =  "gcr.io/google-samples/hello-app:1.0"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}`,
		},
	}
	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			actual, err := FromTerraform(strings.NewReader(c.FileContent))
			if err != nil && !c.WantErr {
				t.Errorf("Unexpected error occurred %s", err)
			}

			if !reflect.DeepEqual(actual, c.ExpectedImages) {
				t.Errorf("Got images list %v, expected %v", actual, c.ExpectedImages)
			}
		})
	}
}
# (2025-03-04)