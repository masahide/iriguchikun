class Iriguchikun < Formula
  desc "tcp/udp/unix domain socket proxy"
  homepage "https://github.com/masahide/iriguchikun"
  url "https://github.com/masahide/iriguchikun/releases/download/v1.1.0/iriguchikun_Darwin_x86_64.tar.gz"
  version "1.1.0"
  sha256 "7168f298a55d7a0edb2f6c5248773a0c595b9f3d07604027ec2c33101b2d296d"

  def install
    bin.install "iriguchikun"
  end

  test do
    system "#{bin}/iriguchikun -v"
  end
end
