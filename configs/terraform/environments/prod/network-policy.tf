resource "kubernetes_network_policy" "prow_cluster_default" {
  metadata {
    name = "prow-cluster-default-network-policy"
  }

  spec {
    pod_selector {}

    // allow outbund connection from any pod
    egress {}
  }
}

resource "kubernetes_network_policy" "trusted_cluster_default" {

  metadata {
    name = "trusted-cluster-default-network-policy"
  }

  spec {
    pod_selector {}

    // allow outbund connection from any pod
    egress {}
  }
}

resource "kubernetes_network_policy" "untrusted_cluster_default" {

  metadata {
    name = "untrusted-cluster-default-network-policy"
  }

  spec {
    pod_selector {}

    // allow outbund connection from any pod
    egress {}
  }
}

resource "kubernetes_network_policy" "trusted_cluster_from_prow" {
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
        ip_block {
          cidr = "10.8.0.0/14"
        }
      }
    }

    policy_types = ["Ingress", "Egress"]
  }
}

resource "kubernetes_network_policy" "untrusted_cluster_from_prow" {
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
