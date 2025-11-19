export type InstanceStatus = 'active' | 'booting' | 'unhealthy' | 'terminated'

export interface Region {
  name: string
  description: string
}

export interface InstanceTypeName {
  name: string
  description: string
}

export interface Instance {
  id: string
  name: string | null
  ip: string | null
  private_ip: string | null
  status: InstanceStatus
  ssh_key_names: string[]
  file_system_names: string[]
  region: Region
  instance_type: InstanceTypeName
  hostname: string | null
  jupyter_token: string | null
  jupyter_url: string | null
}

export interface Specs {
  vcpus: number
  memory_gib: number
  storage_gib: number
  gpus?: number
}

export interface InstanceType {
  name: string
  description: string
  gpu_description: string
  price_cents_per_hour: number
  specs: Specs
  regions_with_capacity_available: Region[]
}

export interface SSHKey {
  id: string
  name: string
  public_key: string
}

export interface FileSystem {
  id: string
  name: string
  mount_point?: string
  created?: string
}

export interface LaunchInstanceRequest {
  instance_type_name: string
  region_name: string
  ssh_key_names: string[]
  file_system_names?: string[]
  quantity?: number
  name?: string
}
