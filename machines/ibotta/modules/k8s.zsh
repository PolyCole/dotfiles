# machines/ibotta/modules/k8s.zsh — Kubernetes cluster context switching

# Commands:
#   kubeconfig-nancy  - Switch to the nancy staging cluster (~/.kube/nancy)
#   kubeconfig-local  - Switch to local cluster (~/.kube/config)

# Nancy (staging) is the default environment.
export KUBECONFIG=~/.kube/nancy

# Sets Kubernetes configuration to the nancy staging cluster.
kubeconfig-nancy(){
  export KUBECONFIG=~/.kube/nancy
}

# Sets Kubernetes configuration to my local machine.
kubeconfig-local() {
  export KUBECONFIG=~/.kube/config
}
