<template>
  <div class="role-container">
    <!-- 表格区域 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色列表</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增
          </el-button>
        </div>
      </template>
      
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="角色名称" width="150" />
        <el-table-column prop="keyword" label="角色关键字" width="150" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column label="操作" fixed="right" width="200">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form ref="formRef" :model="formData" :rules="rules" label-width="100px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色关键字" prop="keyword">
          <el-input v-model="formData.keyword" placeholder="请输入角色关键字" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="formData.description" type="textarea" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="菜单权限" prop="menuIds">
          <el-tree
            ref="menuTreeRef"
            :data="menuTree"
            :props="{ label: 'title', children: 'children' }"
            show-checkbox
            node-key="id"
            :default-checked-keys="formData.menuIds"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="isEdit">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { getRoleList, getRole, createRole, updateRole, deleteRole } from '@/api/role'
import { getMenuList } from '@/api/menu'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const tableData = ref<any[]>([])
const menuTree = ref<any[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const menuTreeRef = ref()

const formData = reactive({
  id: 0,
  name: '',
  keyword: '',
  description: '',
  menuIds: [] as number[],
  apiIds: [] as number[],
  status: 1
})

const dialogTitle = computed(() => isEdit.value ? '编辑角色' : '新增角色')

const rules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  keyword: [{ required: true, message: '请输入角色关键字', trigger: 'blur' }]
}

// 获取数据
const fetchData = async () => {
  loading.value = true
  try {
    const res = await getRoleList({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    tableData.value = res.data.list
    pagination.total = res.data.total
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 获取菜单树
const fetchMenus = async () => {
  try {
    const res = await getMenuList()
    menuTree.value = res.data
  } catch (error) {
    console.error(error)
  }
}

// 新增
const handleAdd = () => {
  isEdit.value = false
  Object.assign(formData, {
    id: 0,
    name: '',
    keyword: '',
    description: '',
    menuIds: [],
    apiIds: [],
    status: 1
  })
  dialogVisible.value = true
}

// 编辑
const handleEdit = async (row: any) => {
  isEdit.value = true
  try {
    const res = await getRole(row.id)
    Object.assign(formData, {
      id: res.data.role.id,
      name: res.data.role.name,
      keyword: res.data.role.keyword,
      description: res.data.role.description,
      menuIds: res.data.menuIds || [],
      apiIds: res.data.apiIds || [],
      status: res.data.role.status
    })
    dialogVisible.value = true
  } catch (error) {
    console.error(error)
  }
}

// 删除
const handleDelete = (row: any) => {
  ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteRole(row.id)
    ElMessage.success('删除成功')
    fetchData()
  })
}

// 提交
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        // 获取选中的菜单
        formData.menuIds = menuTreeRef.value?.getCheckedKeys() || []
        
        if (isEdit.value) {
          await updateRole(formData)
          ElMessage.success('更新成功')
        } else {
          await createRole(formData)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchData()
      } catch (error) {
        console.error(error)
      }
    }
  })
}

onMounted(() => {
  fetchData()
  fetchMenus()
})
</script>

<style scoped lang="scss">
.role-container {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
