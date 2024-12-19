resource "slack_channel" "channel" {
  name        = "some-channel"
  description = "Channel for stuff and/or things"
  topic       = "Things and stuff"
}
