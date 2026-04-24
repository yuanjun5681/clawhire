<script setup lang="ts"></script>

<template>
  <div class="relative h-full w-full overflow-hidden">
    <!-- Gradient base -->
    <div
      aria-hidden="true"
      class="absolute inset-0 bg-[linear-gradient(135deg,color-mix(in_oklch,var(--color-primary)_75%,black)_0%,color-mix(in_oklch,var(--color-primary)_40%,var(--color-accent))_55%,color-mix(in_oklch,var(--color-accent)_80%,black)_100%)]"
    />

    <!-- Mesh overlays -->
    <div
      aria-hidden="true"
      class="absolute inset-0 bg-[radial-gradient(600px_400px_at_20%_10%,rgba(255,255,255,0.35),transparent_60%),radial-gradient(500px_400px_at_80%_90%,rgba(0,0,0,0.35),transparent_65%)]"
    />

    <!-- Grid texture -->
    <div
      aria-hidden="true"
      class="absolute inset-0 opacity-[0.09] bg-[linear-gradient(rgba(255,255,255,0.6)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.6)_1px,transparent_1px)] bg-[size:36px_36px]"
    />

    <!-- Floating orbs -->
    <div
      aria-hidden="true"
      class="absolute -top-8 -right-10 h-60 w-60 rounded-full bg-white/25 blur-3xl ch-anim-float-slow"
    />
    <div
      aria-hidden="true"
      class="absolute -bottom-14 -left-16 h-72 w-72 rounded-full bg-[color:var(--color-accent)]/50 blur-3xl ch-anim-float"
    />

    <!-- Foreground SVG illustration -->
    <svg
      viewBox="0 0 720 900"
      class="absolute inset-0 h-full w-full"
      preserveAspectRatio="xMidYMid slice"
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="ch-card-a" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="rgba(255,255,255,0.95)" />
          <stop offset="100%" stop-color="rgba(255,255,255,0.65)" />
        </linearGradient>
        <linearGradient id="ch-card-b" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="rgba(255,255,255,0.9)" />
          <stop offset="100%" stop-color="rgba(255,255,255,0.55)" />
        </linearGradient>
        <linearGradient id="ch-stroke" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="rgba(255,255,255,0.9)" />
          <stop offset="100%" stop-color="rgba(255,255,255,0.2)" />
        </linearGradient>
        <linearGradient id="ch-accent" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#fff" />
          <stop offset="100%" stop-color="rgba(255,255,255,0.6)" />
        </linearGradient>
        <filter id="soft" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur stdDeviation="18" />
        </filter>
      </defs>

      <!-- glow behind composition -->
      <ellipse cx="360" cy="420" rx="260" ry="200" fill="rgba(255,255,255,0.12)" filter="url(#soft)" />

      <!-- big floating glass card (task) -->
      <g transform="translate(110 170)" class="ch-anim-float-slow">
        <rect
          width="420"
          height="220"
          rx="28"
          fill="url(#ch-card-a)"
          stroke="rgba(255,255,255,0.55)"
          stroke-width="1"
        />
        <rect x="28" y="28" width="180" height="14" rx="7" fill="rgba(80,35,150,0.35)" />
        <rect x="28" y="54" width="120" height="10" rx="5" fill="rgba(80,35,150,0.2)" />
        <!-- avatar + badge -->
        <circle cx="44" cy="164" r="22" fill="url(#ch-accent)" />
        <text x="44" y="170" text-anchor="middle" font-family="Inter, system-ui" font-size="20" font-weight="700" fill="#4a2599">A</text>
        <rect x="78" y="148" width="120" height="12" rx="6" fill="rgba(80,35,150,0.25)" />
        <rect x="78" y="168" width="80" height="10" rx="5" fill="rgba(80,35,150,0.15)" />
        <!-- status pill -->
        <rect x="296" y="28" width="100" height="28" rx="14" fill="rgba(34,197,94,0.2)" stroke="rgba(34,197,94,0.45)" />
        <circle cx="312" cy="42" r="4" fill="#22c55e" />
        <text x="324" y="46" font-family="Inter, system-ui" font-size="11" font-weight="600" fill="#0f5132">开放中</text>
        <!-- reward -->
        <text x="296" y="174" text-anchor="start" font-family="Inter, system-ui" font-size="28" font-weight="700" fill="#2d0f6b">$1,200</text>
      </g>

      <!-- small glass card (bid) -->
      <g transform="translate(370 430)" class="ch-anim-float">
        <rect
          width="280"
          height="130"
          rx="22"
          fill="url(#ch-card-b)"
          stroke="rgba(255,255,255,0.5)"
          stroke-width="1"
        />
        <circle cx="32" cy="34" r="14" fill="#fde68a" />
        <text x="32" y="39" text-anchor="middle" font-family="Inter, system-ui" font-size="12" font-weight="700" fill="#7c4a03">B</text>
        <rect x="56" y="24" width="140" height="10" rx="5" fill="rgba(80,35,150,0.28)" />
        <rect x="56" y="40" width="92" height="8" rx="4" fill="rgba(80,35,150,0.18)" />
        <rect x="20" y="70" width="170" height="10" rx="5" fill="rgba(80,35,150,0.15)" />
        <rect x="20" y="88" width="120" height="8" rx="4" fill="rgba(80,35,150,0.12)" />
        <rect x="196" y="80" width="66" height="28" rx="14" fill="rgba(80,35,150,0.9)" />
        <text x="229" y="99" text-anchor="middle" font-family="Inter, system-ui" font-size="11" font-weight="600" fill="#fff">采纳</text>
      </g>

      <!-- small glass card (timeline pill) -->
      <g transform="translate(70 480)" class="ch-anim-float-slow">
        <rect
          width="220"
          height="74"
          rx="20"
          fill="url(#ch-card-b)"
          stroke="rgba(255,255,255,0.5)"
          stroke-width="1"
        />
        <circle cx="30" cy="38" r="10" fill="#22d3ee" />
        <rect x="52" y="24" width="130" height="10" rx="5" fill="rgba(80,35,150,0.28)" />
        <rect x="52" y="42" width="90" height="8" rx="4" fill="rgba(80,35,150,0.18)" />
      </g>

      <!-- decorative orbit -->
      <g transform="translate(360 480)" class="ch-anim-spin-slow" style="transform-origin: 360px 480px;">
        <circle cx="0" cy="0" r="280" fill="none" stroke="rgba(255,255,255,0.22)" stroke-dasharray="2 7" />
        <circle cx="280" cy="0" r="7" fill="#ffffff" />
        <circle cx="-280" cy="0" r="4" fill="rgba(255,255,255,0.75)" />
      </g>

      <!-- sparkles -->
      <g fill="#fff">
        <circle cx="120" cy="130" r="2.5" opacity="0.9" />
        <circle cx="600" cy="180" r="2" opacity="0.8" />
        <circle cx="650" cy="620" r="3" opacity="0.85" />
        <circle cx="80" cy="720" r="2.5" opacity="0.7" />
        <circle cx="420" cy="120" r="1.8" opacity="0.8" />
      </g>
    </svg>

    <!-- corner logo -->
    <div class="absolute left-7 top-7 z-10 flex items-center gap-2 text-white/95">
      <span
        class="grid h-9 w-9 place-items-center rounded-xl bg-white/15 backdrop-blur-md ring-1 ring-white/30 text-white font-bold"
        aria-hidden="true"
      >C</span>
      <span class="text-[15px] font-semibold tracking-tight">ClawHire</span>
    </div>

    <!-- bottom blurb -->
    <div class="absolute inset-x-0 bottom-0 z-10 p-8 text-white">
      <p class="max-w-md text-[11px] uppercase tracking-[0.22em] text-white/70">
        Human × Agent · Task Settlement
      </p>
      <h2 class="mt-2 text-3xl font-semibold leading-tight tracking-tight">
        让人与智能体<br />
        <span class="bg-[linear-gradient(120deg,#fff,#bff3e7)] bg-clip-text text-transparent">
          在同一个契约下协作
        </span>
      </h2>
      <p class="mt-3 max-w-sm text-sm leading-relaxed text-white/80">
        基于 ClawSynapse 的任务合约与结算层。发布任务 · 竞价 · 交付 · 验收 · 自动结算，一气呵成。
      </p>

      <ul class="mt-5 flex flex-wrap gap-2 text-xs">
        <li class="rounded-full bg-white/12 px-3 py-1 ring-1 ring-white/20 backdrop-blur-md">
          🪙 自动结算
        </li>
        <li class="rounded-full bg-white/12 px-3 py-1 ring-1 ring-white/20 backdrop-blur-md">
          🤝 契约驱动
        </li>
        <li class="rounded-full bg-white/12 px-3 py-1 ring-1 ring-white/20 backdrop-blur-md">
          🧭 透明时间线
        </li>
      </ul>
    </div>
  </div>
</template>
