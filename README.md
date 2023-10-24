Terraform Custom Provider for Alibaba Cloud
===========================================

This Terraform custom provider is designed for own use case scenario.

Supported Versions
------------------

| Terraform version | minimum provider version |maxmimum provider version
| ---- | ---- | ----|
| >= 1.3.x	| 0.1.1	| latest |

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 1.3.x
-	[Go](https://golang.org/doc/install) 1.19 (to build the provider plugin)

Local Installation
------------------

1. Run make file `make install-local-custom-provider` to install the provider under ~/.terraform.d/plugins.

2. The provider source should be change to the path that configured in the *Makefile*:

    ```
    terraform {
      required_providers {
        st-alicloud = {
          source = "example.local/myklst/st-alicloud"
        }
      }
    }

    provider "st-alicloud" {
      region = "cn-hongkong"
    }
    ```

Why Custom Provider
-------------------

This custom provider exists due to some of the resources and data sources in the
official AliCloud Terraform provider may not fulfill the requirements of some
scenario. The reason behind every resources and data sources are stated as below:

### Resources

- **st-alicloud_alidns_gtm_instance**

  The original reason to write this resource is official AliCloud Terraform
  provider does not support creating GTM (Global Traffic Manager) instance on
  AliCloud international account using Terraform. As we developing on the
  resource, we added few more features which are useful in our use case, which
  includes:

    - setting the renewal status to *NotRenewal* when destroying the resource.

    - allowing changing of renewal period and status without recreating the GTM instsance.

- **st-alicloud_alidns_record_weight**

  Official AliCloud Terraform provider does not have the resource to modify DNS
  records weight.

- **st-alicloud_ram_user_group_attachment**

  The official AliCloud Terraform provider's resource
  [*alicloud_ram_group_membership*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/ram_group_membership)
  will remove all other attached users for the target group, which may cause a
  problem where Terraform may delete those users attached outside from Terraform.

- **st-alicloud_ram_policy**

  This resource is designed to handle policy content that exceeds the limit of 6144 characters.
  It provides functionality to create policies by splitting the content into smaller segments that fit within the limit,
  enabling the management and combination of these segments to form the complete policy. Finally, the policy will be attached to the relevant user.

- **st-alicloud_cms_alarm_rule**

  The official AliCloud Terraform provider's resource
  [*alicloud_cms_group_metric_rule*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/cms_group_metric_rule)
  does not support adding alarm rules into application groups based on expression-based creation.

  For namespaces and metric inputs, please refer to: [*Alicloud Alarm Metric List*](https://cms.console.aliyun.com/metric-meta)

- **st-alicloud_alidns_instance**

  The official AliCloud Terraform provider's resource
  [*alicloud_alidns_instance*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/alidns_instance)
  will destroy and create a new instance everytime when upgrading or downgrading.

- **st-alicloud_alidns_domain_attachment**

  The official AliCloud Terraform provider's resource
  [*alicloud_dns_domain_attachment*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/dns_domain_attachment)
  accept input of a list of domains. There will be an issue when upgrading a batch of domains when the existing attachment
  is more than 100 domains. The official resources will first destroy all the domains and re-add the new one together with
  the existing one. The resources will hit timeout during adding of new domains and make some of the domains not re-add back.

- ~~**st-alicloud_cms_system_event_contact_group_attachment**~~

  **Update:**
  - The official AliCloud Terraform provider's resource [*alicloud_cms_event_rule*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/cms_event_rule) is fixed and currently supports the binding of the created system event rule to the contact group.
  - Users are not encouraged to use this resource as it will no longer be maintained.

- **st-alicloud_ddoscoo_webconfig_ssl_attachment**

  This resource is designed to associate a SSL certificate to a website/domain before being added
  into Anti-DDoS and Modify TLS Security Settings as AliCloud Terraform Provider does not support the SSL binding & modify TLS Security Settings operation.

- **st-alicloud_aliadb_resource_group_bind_user**

  This resource is to bind user to a resource group. Binding user to a resource group allow user access to different resource groups.

- **st-alicloud_emr_metric_auto_scaling_rules**

  This resource is designed to create auto scaling rules for AliCloud E-MapReduce cluster as the provider's resource [*alicloud_emrv2_cluster*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/emrv2_cluster)
  does not provide the option to create auto scaling rules for nodes. (Note: Only task nodes are eligible for auto scaling)

- **st-alicloud_ddoscoo_web_ai_protect_config**

  This resource is designed to modify antiddos Web AI Protect Mode Config from Protection to Warning for a website/domain before being added into Anti-DDoS webconfig as AliCloud Terraform Provider does not support
  the modify AI Protect Mode operation.

### Data Sources

- **st-alicloud_ddoscoo_domain_resources**

  - Official AliCloud Terraform provider does not support querying the CNAME of
    AntiDDoS domain resources through
    [*alicloud_ddoscoo_domain_resources*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/data-sources/ddoscoo_domain_resources).

  - Added client_config block to allow overriding the Provider configuration.

- **st-alicloud_ddoscoo_instances**

  Official AliCloud Terraform provider does not support querying the EIP of
  AntiDDoS instance resources through
  [*alicloud_ddoscoo_instances*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/data-sources/ddoscoo_instances).

- **st-alicloud_cdn_domain**

  - Official AliCloud Terraform provider does not have the data source to query
    the CNAME of CDN domain.

  - Added client_config block to allow overriding the Provider configuration.

- **st-alicloud_slb_load_balancers**

  - The tags parameter of AliCloud API
    [*DescribeLoadBalancers*](https://www.alibabacloud.com/help/en/server-load-balancer/latest/describeloadbalancers)
    will return all load balancers when any one of the tags are matched. This may
    be a problem when the user wants to match exactly all given tags, therefore
    this data source will filter once more after listing the load balancers
    from AliCloud API to match all the given tags.

    The example bahaviors of AliCloud API *DescribeLoadBalancers*:

    | Load Balancer   | Tags                                            | Given tags: { "location": "office" "env": "test" }          |
    |-----------------|-------------------------------------------------|-------------------------------------------------------------|
    | load-balancer-A | { "location": "office" "env" : "test" }         | Matched (work as expected)                                  |
    | load-balancer-B | { "location": "office" "env" : "prod" }         | Matched (should not be matched as the `env` is prod)        |

  - Added client_config block to allow overriding the Provider configuration.

- **st-alicloud_cs_user_kubeconfig**

  - The original data source [*alicloud_cs_cluster_credential*](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/data-sources/cs_cluster_credential) has limitations on the following configuration scenarios:
    - Unable to query Kubeconfig of another RAM user using the same Terraform provider as it has different credentials.

  - Added client_config block to allow overriding the Provider configuration.

References
----------

- Website: https://www.terraform.io
- Terraform Plugin Framework: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework
- AliCloud official Terraform provider: https://github.com/aliyun/terraform-provider-alicloud
