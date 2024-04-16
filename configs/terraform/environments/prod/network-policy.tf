resource "kubernetes_network_policy" "prow_cluster_default" {
  provider = kubernetes.prow_k8s_cluster
  metadata {
    name = "prow-cluster-default-network-policy"
  }

  spec {
    pod_selector {}

    // allow outbund connection from any pod
    egress {}

    policy_types = ["Ingress", "Egress"]
  }
}

resource "kubernetes_network_policy" "trusted_cluster_default" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "trusted-cluster-default-network-policy"
  }

  spec {
    pod_selector {}

    // allow outbund connection from any pod
    egress {}

    policy_types = ["Ingress", "Egress"]
  }
}

resource "kubernetes_network_policy" "untrusted_cluster_default" {
  provider = kubernetes.untrusted_workload_k8s_cluster

  metadata {
    name = "untrusted-cluster-default-network-policy"
  }

  spec {
    pod_selector {}

    // allow outbund connection from any pod
    egress {}
    policy_types = ["Ingress", "Egress"]
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
          cidr = var.prow_cluster_ip_range
        }
      }
    }

    policy_types = ["Ingress"]
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
          cidr = var.prow_cluster_ip_range
        }
      }
    }

    policy_types = ["Ingress"]
  }
}

resource "kubernetes_network_policy" "prow_allow_http_events" {
  provider = kubernetes.prow_k8s_cluster

  metadata {
    name = "prow-allow-http-events"
  }

  spec {
    pod_selector {
      match_labels = {
        "app" = "deck",
        "app" = "hook"
      }
    }

    ingress {
      from {
        ip_block {
          cidr = "0.0.0.0/0"
        }
      }
    }

    policy_types = ["Ingress"]
  }
}

resource "kubernetes_network_policy" "hook_to_plugins" {
  provider = kubernetes.prow_k8s_cluster

  metadata {
    name = "hook-to-plugins-network-policy"
  }

  spec {
    pod_selector {
      match_labels = {
        "app" = "automated-approver",
        "app" = "cla-assistant",
      }
    }

    ingress {
      from {
        pod_selector {
          match_labels = {
            "app" = "hook"
          }
        }
      }
    }

    policy_types = ["Ingress"]
  }
}
