workflow "on pull request merge, delete the branch" {
  on = "pull_request"
  resolves = ["branch cleanup"]
}

 action "branch cleanup" {
  uses = "articulate/branch-cleanup-action@master"
  secrets = ["GITHUB_TOKEN"]
}
