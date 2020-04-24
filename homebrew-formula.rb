class Shalm < Formula
  desc "K8s installation tool"
  homepage "https://github.com/wonderix/shalm"
  version "{{version}}"
  if OS.mac?
    url "https://github.com/wonderix/shalm/releases/download/{{version}}/shalm-binary-darwin.tgz"
    sha256 "{{sha256-darwin}}"
  elsif OS.linux?
    url "https://github.com/wonderix/shalm/releases/download/{{version}}/shalm-binary-linux.tgz"
    sha256 "{{sha256-linux}}"
  end

  depends_on :arch => :x86_64

  def install
    bin.install "shalm" => "shalm"
  end

  test do
    system "#{bin}/shalm", "version"
  end
end