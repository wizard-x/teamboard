<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { membersApi } from '@/api'
import { useAuth } from '@/composables/useAuth'
import { useNotifications } from '@/composables/useNotifications'
import type { Member, MemberRole, MemberWithKey } from '@/types'

const { isAdmin, currentMember } = useAuth()
const { notify } = useNotifications()

const members = ref<Member[]>([])
const loading = ref(true)
const showAddForm = ref(false)
const addForm = ref({ name: '', email: '', role: 'member' as MemberRole })
const adding = ref(false)
const newMemberKey = ref<string | null>(null)
const editingMemberId = ref<string | null>(null)
const editForm = ref({ name: '', role: '' as MemberRole })
const saving = ref(false)

onMounted(loadMembers)

async function loadMembers() {
  loading.value = true
  try {
    members.value = await membersApi.list()
  } catch {
    notify('error', 'Failed to load members')
  } finally {
    loading.value = false
  }
}

async function addMember() {
  adding.value = true
  try {
    const m = await membersApi.create(addForm.value)
    newMemberKey.value = m.api_key
    members.value.push({
      id: m.id,
      name: m.name,
      email: m.email,
      role: m.role,
      created_at: m.created_at,
    })
    addForm.value = { name: '', email: '', role: 'member' }
    notify('success', 'Member added')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to add member')
  } finally {
    adding.value = false
  }
}

function startEdit(member: Member) {
  editingMemberId.value = member.id
  editForm.value = { name: member.name, role: member.role }
}

async function saveEdit() {
  if (!editingMemberId.value) return
  saving.value = true
  try {
    const updated = await membersApi.update(editingMemberId.value, editForm.value)
    const idx = members.value.findIndex((m) => m.id === editingMemberId.value)
    if (idx !== -1) members.value[idx] = updated
    editingMemberId.value = null
    notify('success', 'Member updated')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to update member')
  } finally {
    saving.value = false
  }
}

async function removeMember(id: string) {
  if (!confirm('Remove this member?')) return
  try {
    await membersApi.remove(id)
    members.value = members.value.filter((m) => m.id !== id)
    notify('success', 'Member removed')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to remove member')
  }
}

async function regenerateKey(id: string) {
  if (!confirm('Regenerate API key? The old key will stop working immediately.')) return
  try {
    const result = await membersApi.regenerateKey(id)
    newMemberKey.value = result.api_key
    notify('success', 'API key regenerated')
  } catch (e: any) {
    notify('error', e.response?.data?.error?.message || 'Failed to regenerate key')
  }
}

function copyKey() {
  if (newMemberKey.value) {
    navigator.clipboard.writeText(newMemberKey.value)
    notify('success', 'API key copied to clipboard!')
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString()
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1>Team Members</h1>
      <button v-if="isAdmin" class="btn btn-primary" @click="showAddForm = true">+ Add Member</button>
    </div>

    <!-- New member key display -->
    <div v-if="newMemberKey" class="key-banner card" style="margin-bottom: 16px">
      <p><strong>New API Key generated:</strong></p>
      <div class="key-row">
        <code>{{ newMemberKey }}</code>
        <button class="btn btn-sm btn-secondary" @click="copyKey">Copy</button>
        <button class="btn btn-sm btn-secondary" @click="newMemberKey = null">Dismiss</button>
      </div>
      <p class="key-warning">Save this key now — it won't be shown again.</p>
    </div>

    <div v-if="loading" class="loading-container">
      <span class="spinner" />
      Loading members...
    </div>

    <!-- Add Member Form -->
    <div v-if="showAddForm" class="card" style="margin-bottom: 16px">
      <h3 style="margin-bottom: 12px">Add Member</h3>
      <form @submit.prevent="addMember" style="display: flex; gap: 12px; align-items: flex-end; flex-wrap: wrap">
        <div class="form-group" style="flex: 1; min-width: 150px; margin-bottom: 0">
          <label>Name</label>
          <input v-model="addForm.name" class="form-control" placeholder="Jane Doe" maxlength="100" />
        </div>
        <div class="form-group" style="flex: 1; min-width: 200px; margin-bottom: 0">
          <label>Email</label>
          <input v-model="addForm.email" type="email" class="form-control" placeholder="jane@example.com" maxlength="255" />
        </div>
        <div class="form-group" style="min-width: 120px; margin-bottom: 0">
          <label>Role</label>
          <select v-model="addForm.role" class="form-control">
            <option value="member">Member</option>
            <option value="admin">Admin</option>
          </select>
        </div>
        <div style="display: flex; gap: 8px">
          <button type="submit" class="btn btn-primary" :disabled="adding">Add</button>
          <button type="button" class="btn btn-secondary" @click="showAddForm = false">Cancel</button>
        </div>
      </form>
    </div>

    <!-- Members Table -->
    <div v-if="!loading" class="card">
      <table class="members-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
            <th>Joined</th>
            <th v-if="isAdmin">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="member in members" :key="member.id">
            <template v-if="editingMemberId === member.id">
              <td>
                <input v-model="editForm.name" class="form-control" style="min-width: 120px" />
              </td>
              <td>{{ member.email }}</td>
              <td>
                <select v-model="editForm.role" class="form-control" style="min-width: 100px">
                  <option value="member">Member</option>
                  <option value="admin">Admin</option>
                </select>
              </td>
              <td>{{ formatDate(member.created_at) }}</td>
              <td>
                <div style="display: flex; gap: 4px">
                  <button class="btn btn-sm btn-primary" :disabled="saving" @click="saveEdit">Save</button>
                  <button class="btn btn-sm btn-secondary" @click="editingMemberId = null">Cancel</button>
                </div>
              </td>
            </template>
            <template v-else>
              <td>
                <strong>{{ member.name }}</strong>
                <span v-if="currentMember?.id === member.id" style="color: var(--color-text-muted); font-size: 12px"> (you)</span>
              </td>
              <td>{{ member.email }}</td>
              <td><span :class="['badge', `badge-${member.role}`]">{{ member.role }}</span></td>
              <td>{{ formatDate(member.created_at) }}</td>
              <td v-if="isAdmin && currentMember?.id !== member.id">
                <div style="display: flex; gap: 4px">
                  <button class="btn btn-sm btn-secondary" @click="startEdit(member)">Edit</button>
                  <button class="btn btn-sm btn-secondary" @click="regenerateKey(member.id)">Regen Key</button>
                  <button class="btn btn-sm btn-danger" @click="removeMember(member.id)">Remove</button>
                </div>
              </td>
              <td v-else-if="isAdmin" />
            </template>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.members-table {
  width: 100%;
  border-collapse: collapse;
}

.members-table th,
.members-table td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  font-size: 14px;
}

.members-table th {
  font-weight: 600;
  color: var(--color-text-muted);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.key-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.key-row code {
  flex: 1;
  padding: 8px 12px;
  background: var(--color-bg);
  border-radius: var(--radius);
  word-break: break-all;
  font-size: 13px;
}

.key-warning {
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-warning);
}
</style>
