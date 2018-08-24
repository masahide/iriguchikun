class Iriguchikun < Formula
  desc "tcp/udp/unix domain socket proxy"
  homepage "https://github.com/masahide/iriguchikun"
  url "https://github.com/masahide/iriguchikun/releases/download/v1.2.0/iriguchikun_Darwin_x86_64.tar.gz"
  version "1.2.0"
  sha256 "859bd818f01d725448557c0e5d31ad0197172795be4203f1ff972c43cf055050"

  def install
    bin.install "iriguchikun"
  end

  test do
    system "#{bin}/iriguchikun -v"
  end
end
