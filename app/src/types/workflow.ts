// ComfyUI Workflow Types

export interface ComfyUIWorkflow {
  id: string
  name: string
  description: string
  category: 'image' | 'video' | 'upscaling' | 'custom'
  icon: string
  thumbnail?: string
  workflow_json: string // JSON stringified ComfyUI workflow
  parameters: WorkflowParameter[]
  outputs: WorkflowOutput[]
}

export interface WorkflowParameter {
  id: string
  name: string
  type: 'text' | 'number' | 'image' | 'select' | 'checkbox'
  description: string
  required: boolean
  default_value?: string | number | boolean
  options?: string[] // For select type
  min?: number // For number type
  max?: number // For number type
  node_id?: string // ComfyUI node ID to update
  field_name?: string // Field in the node to update
}

export interface WorkflowOutput {
  type: 'image' | 'video' | 'file'
  name: string
  format: string
}

export interface WorkflowExecution {
  id: string
  workflow_id: string
  status: 'queued' | 'running' | 'completed' | 'failed'
  progress: number
  prompt_id?: string
  queue_position?: number
  current_node?: string
  error?: string
  started_at?: string
  completed_at?: string
  outputs?: WorkflowExecutionOutput[]
}

export interface WorkflowExecutionOutput {
  filename: string
  subfolder: string
  type: string
  url: string
}

export interface ComfyUIStatus {
  connected: boolean
  queue_remaining: number
  queue_running: number
  system_stats?: {
    devices: DeviceStats[]
  }
}

export interface DeviceStats {
  name: string
  type: string
  vram_total: number
  vram_free: number
  torch_vram_total: number
  torch_vram_free: number
}

export interface QueuePromptRequest {
  workflow_json: string
  parameters: Record<string, any>
}

export interface QueuePromptResponse {
  prompt_id: string
  number: number
}

export interface HistoryItem {
  prompt: any[]
  outputs: Record<string, any>
  status: {
    status_str: string
    completed: boolean
    messages: any[]
  }
}
