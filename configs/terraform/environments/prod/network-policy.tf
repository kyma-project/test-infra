resource "kubernetes_network_policy" "trusted_cluster_to_others" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name      = "trusted-egress-network-policy"
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

      from {
        ip_block {
          cidr = "10.0.0.0/14"
        }
      }

      from {
        ip_block {
          cidr = "10.109.0.0/19"
        }
      }
    }

    policy_types = ["Ingress", "Egress"]
  }
}
