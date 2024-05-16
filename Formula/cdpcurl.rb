class Cdpcurl < Formula
  desc "curl with CDP API key"
  homepage "https://github.com/dimei-BT/cdpcurl"
  url "file:///Users/dimei/Downloads/cdpcurl-0.0.1.tar.gz"
  sha256 "a894b2f9bffa020b7cbfa877fed9aa25fb6529c85376590ce0eb421f4fcd40ed"

  def install
    system "go", "build", "-o", "#{bin}/cdpcurl", "."
  end

  test do
    system "#{bin}/cdpcurl", "--version"
  end
end
