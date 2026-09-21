# speckeep formula — lives in the tap repo github.com/bzdvdn/homebrew-speckeep
# at Formula/speckeep.rb. Publish via:
#
#   brew tap bzdvdn/speckeep
#   brew install bzdvdn/speckeep/speckeep
class Speckeep < Formula
  desc "Lightweight Spec-Driven Development kit for development agents and humans"
  homepage "https://github.com/bzdvdn/speckeep"
  url "https://github.com/bzdvdn/speckeep/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "7a6bc4730a07fe9573513d75a47f1deeed504edbac66c1ca02ea475f88bd167b"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath",
           "-ldflags", "-X speckeep/src/internal/cli.Version=v#{version}",
           "-o", bin/"speckeep", "./src/cmd/speckeep"
  end

  test do
    assert_match(/v#{version}/, shell_output("#{bin}/speckeep --version"))
  end
end