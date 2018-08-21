package ubuntu

const userdataNodeTemplate = `#cloud-config
hostname: {{ .MachineSpec.Name }}

ssh_pwauth: no

{{- if ne (len .ProviderConfig.SSHPublicKeys) 0 }}
ssh_authorized_keys:
{{- range .ProviderConfig.SSHPublicKeys }}
  - "{{ . }}"
{{- end }}
{{- end }}

write_files:
- path: "/etc/sysctl.d/k8s.conf"
  content: |
    net.bridge.bridge-nf-call-ip6tables = 1
    net.bridge.bridge-nf-call-iptables = 1
    kernel.panic_on_oops = 1
    kernel.panic = 10
    vm.overcommit_memory = 1

- path: "/etc/yum.repos.d/kubernetes.repo"
  content: |
    [kubernetes]
    name=Kubernetes
    baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-$basearch
    enabled=1
    gpgcheck=1
    repo_gpgcheck=1
    gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg

- path: /etc/sysconfig/selinux
  content: |
    # This file controls the state of SELinux on the system.
    # SELINUX= can take one of these three values:
    #     enforcing - SELinux security policy is enforced.
    #     permissive - SELinux prints warnings instead of enforcing.
    #     disabled - No SELinux policy is loaded.
    SELINUX=permissive
    # SELINUXTYPE= can take one of three two values:
    #     targeted - Targeted processes are protected,
    #     minimum - Modification of targeted policy. Only selected processes are protected.
    #     mls - Multi Level Security protection.
    SELINUXTYPE=targeted

- path: "/etc/sysconfig/kubelet-overwrite"
  content: |
    KUBELET_DNS_ARGS=
    KUBELET_EXTRA_ARGS=--authentication-token-webhook=true \
      --hostname-override={{ .MachineSpec.Name }} \
      --read-only-port=0 \
      --protect-kernel-defaults=true \
      --cluster-domain=cluster.local

{{- if semverCompare "<1.11.0" .KubeletVersion }}
- path: "/etc/systemd/system/kubelet.service.d/20-extra.conf"
  content: |
    [Service]
    EnvironmentFile=/etc/sysconfig/kubelet
{{- end }}

- path: "/usr/local/bin/setup"
  permissions: "0777"
  content: |
    #!/bin/bash
    set -xeuo pipefail
    setenforce 0 || true
    sysctl --system

    # There is a dependency issue in the rpm repo for 1.8, if the cni package is not explicitly
    # specified, installation of the kube packages fails
    export CNI_PKG=''
    {{- if semverCompare "=1.8.X" .KubeletVersion }}
    export CNI_PKG='kubernetes-cni-0.5.1-1'
    {{- end }}

    yum install -y {{ .DockerPackageName }} \
      kubelet-{{ .KubeletVersion }} \
      kubeadm-{{ .KubeletVersion }} \
      ebtables \
      ethtool \
      nfs-utils \
      bash-completion \
      sudo \
      ${CNI_PKG}

    cp /etc/sysconfig/kubelet-overwrite /etc/sysconfig/kubelet

    systemctl enable --now docker
    systemctl enable --now kubelet

    kubeadm join \
      --token {{ .BoostrapToken }} \
      -discovery-token-unsafe-skip-ca-verification \
      {{- if semverCompare ">=1.9.X" .KubeletVersion }}
      --ignore-preflight-errors=CRI \
      {{- end }}
      {{ .ServerAddr }}

- path: "/usr/local/bin/supervise.sh"
  permissions: "0777"
  content: |
    #!/bin/bash
    set -xeuo pipefail
    while ! "$@"; do
      sleep 1
    done

- path: "/etc/systemd/system/setup.service"
  content: |
    [Install]
    WantedBy=multi-user.target

    [Unit]
    Requires=network-online.target
    After=network-online.target

    [Service]
    Type=oneshot
    RemainAfterExit=true
    ExecStart=/usr/local/bin/supervise.sh /usr/local/bin/setup

runcmd:
- systemctl enable --now setup.service
`
