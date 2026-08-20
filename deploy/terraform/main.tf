terraform {
  required_version = ">= 1.5.0"

  required_providers {
    kind = {
      source  = "tehcyx/kind"
      version = "~> 0.4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.30.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.13.0" # Add Helm
    }
  }
}

provider "kind" {}

provider "kubernetes" {
  host                   = kind_cluster.local_k8s.endpoint
  client_certificate     = kind_cluster.local_k8s.client_certificate
  client_key             = kind_cluster.local_k8s.client_key
  cluster_ca_certificate = kind_cluster.local_k8s.cluster_ca_certificate
}

provider "helm" {
  kubernetes {
    host                   = kind_cluster.local_k8s.endpoint
    client_certificate     = kind_cluster.local_k8s.client_certificate
    client_key             = kind_cluster.local_k8s.client_key
    cluster_ca_certificate = kind_cluster.local_k8s.cluster_ca_certificate
  }
}
