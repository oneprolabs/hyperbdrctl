package i18n

const bootConfigApplyAliyunBlockEN = `Create or update a boot configuration from cloud account ` + "`<account_id>`" + `; provider ` + "`aliyun`" + `, storage type ` + "`block`" + `.

Parameter Sources:
  --storage-id string
    Use the cloud sync gateway ID, usually from:
      hyperbdrctl cloud-sync-gateway list

Prepare the inputs in this order:
  1. Confirm the base resources:
    hyperbdrctl host list --status host_register_done
    hyperbdrctl cloud-account detail --id <account_id>
    hyperbdrctl cloud-sync-gateway list

  2. Query cloud resources:
    Query zones first:
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --fetch-res zones

    Then query flavors:
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --fetch-res flavors \
        --flavor-vcpus 2 \
        --flavor-ram 4

    Then query volume types:
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --flavor-id <flavor_id> \
        --purpose make_hg \
        --image_type=system \
        --fetch-res system_volume_types,volume_types

    Then query networks and subnets:
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --fetch-res networks,subnets

    Then query security groups:
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --network-id <network_id> \
        --fetch-res security_groups

The minimum apply command is:
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id>

To continue overriding metadata, add:
  --set key=value
  --set-json key=<json>

The override order is:
  --set-json < --set < explicit and dynamic flags

To inspect the final request body without sending a write request, add:
  --preview-request`

const bootConfigApplyAliyunObjectEN = `Create or update a boot configuration from cloud account ` + "`<account_id>`" + `; provider ` + "`aliyun`" + `, storage type ` + "`object`" + `.

Parameter Sources:
  --storage-id string
    Use the target object storage ID, usually from ` + "`hyperbdrctl oss list`" + `.
  --system-volume-type-id string
    Use the system-disk volume type ID.
  --volume-type-id string
    Use the data-disk volume type ID.

Prepare the inputs in this order:
  1. Confirm the base resources with ` + "`host list`" + `, ` + "`cloud-account detail`" + `, and ` + "`oss list`" + `.
  2. Query zones; flavors and OS types; volume types; networks; then subnets and security groups:
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res zones
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --zone-id <zone_id> --fetch-res flavors,os_types --flavor-vcpus 2 --flavor-ram 4
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --zone-id <zone_id> --flavor-id <flavor_id> --fetch-res system_volume_types,volume_types
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res networks
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --zone-id <zone_id> --network-id <network_id> --fetch-res subnets,security_groups

The minimum apply command is:
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id> \
    --bandwidth-size 100

To continue overriding metadata, add ` + "`--set key=value`" + ` or ` + "`--set-json key=<json>`" + `.
The override order is ` + "`--set-json < --set < explicit and dynamic flags`" + `.
To inspect the final request body without sending a write request, add ` + "`--preview-request`" + `.`

const bootConfigApplyHuaweiObjectEN = `Create or update a boot configuration from cloud account ` + "`<account_id>`" + `; provider ` + "`huawei`" + `, storage type ` + "`object`" + `.

Parameter Sources:
  --storage-id string
    Use the target object storage ID, usually from ` + "`hyperbdrctl oss list`" + `.
  --system-volume-type-id string
    Use the system-disk volume type ID.
  --volume-type-id string
    Use the data-disk volume type ID.

Prepare the inputs in this order:
  1. Confirm the base resources with ` + "`host list`" + `, ` + "`cloud-account detail`" + `, and ` + "`oss list`" + `.
  2. Query regions and zones; flavors and OS types; volume types; networks; then subnets and security groups:
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res regions,zones
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --zone-id <zone_id> --fetch-res flavors,os_types --flavor-vcpus 2 --flavor-ram 4
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --zone-id <zone_id> --fetch-res system_volume_types,volume_types
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res networks
    hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res subnets,security_groups

The minimum apply command is:
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id> \
    --bandwidth-size 100

To continue overriding metadata, add ` + "`--set key=value`" + ` or ` + "`--set-json key=<json>`" + `.
The override order is ` + "`--set-json < --set < explicit and dynamic flags`" + `.
To inspect the final request body without sending a write request, add ` + "`--preview-request`" + `.`

const bootConfigApplyOpenstackBlockEN = `Create or update a boot configuration from cloud account ` + "`<account_id>`" + `; provider ` + "`openstack`" + `, storage type ` + "`block`" + `.

Parameter Sources:
  --storage-id string
    Use the cloud sync gateway ID, usually from ` + "`hyperbdrctl cloud-sync-gateway list`" + `.
  --volume-type-id string
    Use the volume type ID.

Confirm the host, account, and gateway first. Then query:
  regions,compute_zones,projects
  flavors,os_types
  images
  volume_types
  networks
  subnets,security_groups

Use ` + "`hyperbdrctl cloud-resource fetch --cloud-account-id <account_id>`" + ` with the selected region, project, compute zone, flavor, and network context for those queries.

The minimum apply command is:
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --compute-zone-id <compute_zone_id> \
    --project-id <project_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id>

To continue overriding metadata, add ` + "`--set key=value`" + ` or ` + "`--set-json key=<json>`" + `.
The override order is ` + "`--set-json < --set < explicit and dynamic flags`" + `.
To inspect the final request body without sending a write request, add ` + "`--preview-request`" + `.`

const bootConfigApplyOpenstackObjectEN = `Create or update a boot configuration from cloud account ` + "`<account_id>`" + `; provider ` + "`openstack`" + `, storage type ` + "`object`" + `.

Parameter Sources:
  --storage-id string
    Use the target object storage ID, usually from ` + "`hyperbdrctl oss list`" + `.
  --system-volume-type-id string
    Use the system-disk volume type ID.
  --volume-type-id string
    Use the data-disk volume type ID.

Confirm the host, account, and object storage first. Then query:
  regions,compute_zones,projects
  flavors,os_types
  system_volume_types,volume_types
  networks
  subnets,security_groups

Use ` + "`hyperbdrctl cloud-resource fetch --cloud-account-id <account_id>`" + ` with the selected region, project, compute zone, flavor, and network context for those queries.

The minimum apply command is:
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id>

To continue overriding metadata, add ` + "`--set key=value`" + ` or ` + "`--set-json key=<json>`" + `.
The override order is ` + "`--set-json < --set < explicit and dynamic flags`" + `.
To inspect the final request body without sending a write request, add ` + "`--preview-request`" + `.`
