resource "kind_cluster" "local_k8s" {
  name           = "cloud-cinema-cluster"
  node_image     = "kindest/node:v1.30.0"
  wait_for_ready = true

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    node {
      role = "control-plane"

      extra_port_mappings {
        container_port = 80
        host_port      = 9080
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 443
        host_port      = 9443
        protocol       = "TCP"
      }

      kubeadm_config_patches = [
        "kind: InitConfiguration\nnodeRegistration:\n  kubeletExtraArgs:\n    node-labels: \"ingress-ready=true\"\n"
      ]
    }
  }
}
