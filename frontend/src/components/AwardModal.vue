<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import UiModal from '@/components/ui/UiModal.vue'
import { formatReward } from '@/utils/format'
import type { AwardTaskInput } from '@/api/tasks'
import type { Bid, TaskDetail } from '@/types'

const props = defineProps<{
  open: boolean
  task: TaskDetail
  bids: Bid[]
}>()

const emit = defineEmits<{
  close: []
  award: [payload: AwardTaskInput]
}>()

const pendingBids = computed(() => props.bids.filter((b) => b.status === 'pending'))

const selectedBidId = ref<string | null>(null)
const amountInput = ref('')

watch(
  () => props.open,
  (open) => {
    if (open) {
      selectedBidId.value = null
      amountInput.value = String(props.task.reward.amount)
    }
  },
)

function selectBid(bidId: string) {
  selectedBidId.value = bidId
  const bid = pendingBids.value.find((b) => b.bidId === bidId)
  if (bid) amountInput.value = String(bid.price)
}

function clearBidSelection() {
  selectedBidId.value = null
  amountInput.value = String(props.task.reward.amount)
}

const selectedBid = computed(() =>
  pendingBids.value.find((b) => b.bidId === selectedBidId.value) ?? null,
)

const canConfirm = computed(() => {
  const amount = Number(amountInput.value)
  return selectedBid.value !== null && Number.isFinite(amount) && amount > 0
})

function nextId(prefix: string) {
  return `${prefix}_${crypto.randomUUID?.()?.slice(0, 8) ?? Date.now()}`
}

function confirm() {
  if (!canConfirm.value || !selectedBid.value) return
  emit('award', {
    contractId: nextId('contract'),
    executorId: selectedBid.value.executor.id,
    agreedReward: {
      amount: Number(amountInput.value),
      currency: props.task.reward.currency || 'USD',
    },
  })
}
</script>

<template>
  <UiModal
    :open="open"
    title="指派执行方"
    description="从报价列表中选择执行方进行指派。"
    size="sm"
    @close="$emit('close')"
  >
    <div class="space-y-5">
      <div v-if="pendingBids.length > 0" class="space-y-2">
        <p class="text-[11px] font-semibold uppercase tracking-widest text-base-content/50">
          当前报价
        </p>
        <ul class="space-y-2">
          <li
            v-for="bid in pendingBids"
            :key="bid.bidId"
            :class="[
              'cursor-pointer rounded-field border p-3 transition',
              selectedBidId === bid.bidId
                ? 'border-primary bg-primary/8 ring-1 ring-primary/40'
                : 'border-base-300/70 bg-base-200/30 hover:border-primary/40 hover:bg-base-200/60',
            ]"
            @click="selectedBidId === bid.bidId ? clearBidSelection() : selectBid(bid.bidId)"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex items-center gap-1.5">
                  <span
                    :class="[
                      'inline-flex h-4 w-4 shrink-0 items-center justify-center rounded-full text-[9px] font-bold',
                      bid.executor.kind === 'agent'
                        ? 'bg-secondary/15 text-secondary'
                        : 'bg-primary/15 text-primary',
                    ]"
                  >
                    {{ bid.executor.kind === 'agent' ? 'A' : 'H' }}
                  </span>
                  <span class="truncate text-sm font-medium text-base-content">
                    {{ bid.executor.name }}
                  </span>
                </div>
                <p
                  v-if="bid.proposal"
                  class="mt-1 line-clamp-2 text-xs text-base-content/60"
                >
                  {{ bid.proposal }}
                </p>
              </div>
              <span class="shrink-0 text-sm font-semibold text-base-content">
                {{ formatReward(bid.price, bid.currency) }}
              </span>
            </div>
          </li>
        </ul>
      </div>

      <div v-else class="rounded-field border border-dashed border-base-300 py-4 text-center text-xs text-base-content/45">
        暂无报价，无法指派。
      </div>

      <div v-if="pendingBids.length > 0" class="space-y-1.5">
        <label class="text-[11px] font-semibold uppercase tracking-widest text-base-content/50">
          约定金额（{{ task.reward.currency || 'USD' }}）
        </label>
        <input
          v-model="amountInput"
          type="number"
          min="0"
          step="any"
          class="input input-bordered w-full text-sm"
        />
      </div>
    </div>

    <template #footer>
      <button class="btn btn-ghost btn-sm" @click="$emit('close')">取消</button>
      <button
        class="btn btn-primary btn-sm"
        :disabled="!canConfirm"
        @click="confirm"
      >
        确认指派
      </button>
    </template>
  </UiModal>
</template>
