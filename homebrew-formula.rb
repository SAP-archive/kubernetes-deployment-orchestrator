class KuberneteDeploymentOrchestrator < Formula
  desc "K8s installation tool"
  homepage "https://github.com/sap/kubernetes-deployment-orchestrator"
  version "{{version}}"
  if OS.mac?
    url "https://github.com/sap/kubernetes-deployment-orchestrator/releases/download/{{version}}/kdo-binary-darwin.tgz"
    sha256 "{{sha256-darwin}}"
  elsif OS.linux?
    url "https://github.com/sap/kubernetes-deployment-orchestrator/releases/download/{{version}}/kdo-binary-linux.tgz"
    sha256 "{{sha256-linux}}"
  end

  depends_on :arch => :x86_64

  def install
    bin.install "kdo" => "kdo"
  end

  test do
    system "#{bin}/kdo", "version"
  end
end