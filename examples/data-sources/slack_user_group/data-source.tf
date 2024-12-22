data "slack_usergroup" "usergroup_by_id" {
  id = "SXXXXXXXX"
}


data "slack_usergroup" "usergroup_by_handle" {
  handle = "my-group"
}
