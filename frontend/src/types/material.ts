// 学习资料（xmind 笔记分享）
export interface StudyMaterialResp {
  id: number
  title: string
  description: string
  subject_id: number
  subject_name: string
  cover_url: string
  file_url: string
  file_name: string
  file_size: number
  download_count: number
  view_count: number
  created_at: string
}

// 思维导图节点（children 为空表示叶子）
export interface MindMapNode {
  title: string
  children?: MindMapNode[]
}
