class Iriguchikun < Formula
  desc "tcp/udp/unix domain socket proxy"
  homepage "https://github.com/masahide/iriguchikun"
  url "https://github.com/masahide/iriguchikun/releases/download/v1.2.2/iriguchikun_Darwin_x86_64.tar.gz"
  version "1.2.2"
  sha256 "a2392f7874fa5ce7ec18367211de9aa0ef29b3706b6db4c6ea208e17570a9590"

  def install
    bin.install "iriguchikun"
  end

  test do
    system "#{bin}/iriguchikun -v"
  end
end
