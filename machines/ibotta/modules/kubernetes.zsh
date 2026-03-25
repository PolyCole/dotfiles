# machines/ibotta/modules/kubernetes.zsh — Kubernetes cluster context switching

# Commands:
#   kubeconfig-prod   - Switch to production cluster (~/.kube/barrel)
#   kubeconfig-stage  - Switch to staging cluster (~/.kube/apollo)
#   kubeconfig-local  - Switch to local cluster (~/.kube/config)

# Staging is the default environment.
export KUBECONFIG=~/.kube/apollo

# Sets Kubernetes configuration to the production cluster.
kubeconfig-prod(){
  export KUBECONFIG=~/.kube/barrel
}

# Sets Kubernetes configuration to the staging cluster.
kubeconfig-stage(){
  export KUBECONFIG=~/.kube/apollo
}

# Sets Kubernetes configuration to my local machine.
kubeconfig-local() {
  export KUBECONFIG=~/.kube/config
}
