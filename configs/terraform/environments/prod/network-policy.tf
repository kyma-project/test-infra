resource "kubernetes_network_policy" "trusted_cluster_to_others" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name      = "trusted-to-prow-policy"
    namespace = "default"
  }

  spec {
    // allow any pods
    pod_selector {}

    ingress {
      from {
        // allow all GKE clusters
        ip_block {
          cidr = "10.8.0.0/14"
        }
      }
    }

    policy_types = ["Ingress", "Egress"]
  }
}

resource "kubernetes_network_policy" "untrusted_cluster_to_others" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name      = "untrusted-to-prow-policy"
    namespace = "default"
  }

  spec {
    // allow any pods
    pod_selector {}

    ingress {
      from {
        // allow all GKE clusters
        ip_block {
          cidr = "10.8.0.0/14"
        }
      }
    }

    policy_types = ["Ingress", "Egress"]
  }
}
