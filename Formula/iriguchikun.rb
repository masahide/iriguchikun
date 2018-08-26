class Iriguchikun < Formula
  desc "tcp/udp/unix domain socket proxy"
  homepage "https://github.com/masahide/iriguchikun"
  url "https://github.com/masahide/iriguchikun/releases/download/v1.2.1/iriguchikun_Darwin_x86_64.tar.gz"
  version "1.2.1"
  sha256 "10e37ffa5474b7f529f9239e04d73c9ad62bcf05805e5c7df0c5b6db9c96d47b"

  def install
    bin.install "iriguchikun"
  end

  test do
    system "#{bin}/iriguchikun -v"
  end
end
