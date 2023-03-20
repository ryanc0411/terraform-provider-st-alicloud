provider "st-alicloud" {
  alias  = "antiddos-instance"
  region = "ap-southeast-1"
}

data "st-alicloud_ddoscoo_instances" "ins" {
  provider     = st-alicloud.antiddos-instance
  ids          = ["id1", "id2"]
  remark_regex = "^example-remark"
}

output "alicloud_ddoscoo_instances" {
  value = data.st-alicloud_ddoscoo_instances.ins
}
