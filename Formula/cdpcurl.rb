class Cdpcurl < Formula
  desc "curl with CDP API key"
  homepage "https://github.com/dimei-BT/cdpcurl"
  url "https://github.com/dimei-BT/cdpcurl/archive/cdpcurl-v0.0.1.tar.gz"
  sha256 "028e4a22226baae713232474f9c5ae02b361f9b63564ac3c838f8b53162cf176"

  def install
    system "go", "build", "-o", "#{bin}/cdpcurl", "."
  end

  test do
    system "#{bin}/cdpcurl", "--version"
  end
end
