resource "st-alicloud_ddoscoo_web_ai_protect_config" "test" {
  enabled = true
  domain  = "api.xxxx.com"
  mode    = "warning"
  level   = "normal"
}
