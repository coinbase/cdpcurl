class Cdpcurl < Formula
  desc "curl with CDP API key"
  homepage "https://github.com/dimei-BT/cdpcurl"
  url "file:///Users/dimei/Downloads/cdpcurl-0.0.1.tar.gz"
  sha256 "d3d4160d83b7f44e49c919ff3d3daf8e2ac6001b9054ab461c45d5d7ca8fa57e"
  license "MIT"

  depends_on "go" => :build

  def install
    cd "cdpcurl" do
      system "go", "build", "-o", "#{bin}/cdpcurl", "."
    end
  end

  test do
    assert_match "cdpcurl version", shell_output("#{bin}/cdpcurl --version")
  end
end