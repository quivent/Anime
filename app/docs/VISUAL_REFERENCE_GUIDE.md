# Lambda Cloud Visual Reference Guide

**Quick reference for expected visual appearance of all UI elements**

---

## Color Palette

### Primary Colors
- **Electric Blue:** `#3b82f6` (primary actions, focus states)
- **Mint Green:** `#10b981` (success, active status)
- **Sunset Orange:** `#f97316` (warnings, destructive actions)
- **Sakura Pink:** `#ec4899` (selection, special states)
- **Neon:** `#a855f7` (Jupyter, special features)

### Status Colors
- **Active:** Mint/Green (#10b981)
- **Booting:** Electric Blue (#3b82f6)
- **Unhealthy:** Sunset Orange (#f97316)
- **Terminated:** Gray (#6b7280)
- **Terminating:** Red (#ef4444)

### UI Grays
- **Background:** Gray 950 (#030712)
- **Card Background:** Gray 900 (#111827) with 50% opacity
- **Border:** Gray 800 (#1f2937)
- **Text Primary:** Gray 100 (#f3f4f6)
- **Text Secondary:** Gray 400 (#9ca3af)

---

## Typography

### Fonts
- **UI Text:** System default (SF Pro on macOS)
- **Code/Monospace:** Monospace font family
- **Special:** Gradient text effects on instance names

### Sizes
- **Page Title:** text-3xl (1.875rem / 30px)
- **Section Headers:** text-xl (1.25rem / 20px)
- **Card Title:** text-lg (1.125rem / 18px)
- **Body Text:** text-sm (0.875rem / 14px)
- **Small Text:** text-xs (0.75rem / 12px)

---

## Button States

### Primary Action Button (Launch, Monitor)
```
Default:
  bg-electric-500/20
  border: border-electric-500/50
  text: text-electric-400

Hover:
  bg-electric-500/30
  (border and text same)

Disabled:
  opacity-50
  cursor-not-allowed

With Glow:
  anime-glow class adds subtle shadow
```

### Secondary Button (Refresh, Close)
```
Default:
  bg-gray-800/50
  border: border-gray-700
  text: text-gray-300

Hover:
  bg-gray-700/50
  (border and text same)
```

### Destructive Button (Terminate)
```
Default:
  bg-sunset-500/20
  border: border-sunset-500/50
  text: text-sunset-400

Hover:
  bg-sunset-500/30
```

### Success Button (Copy, Active)
```
Default:
  bg-mint-500/20
  border: border-mint-500/50
  text: text-mint-400

Hover:
  bg-mint-500/30
```

### Selection Button (Unselected)
```
Default:
  bg-gray-800/50
  border: 1px solid border-gray-700
  text: text-gray-300

Hover:
  border: border-electric-500/30
```

### Selection Button (Selected)
```
Selected:
  bg-electric-500/20
  border: 2px solid border-electric-500/50
  text: text-electric-400
  ring: 2px ring-electric-500/30 (on instance types)
```

---

## Status Badges

### Active Instance
```
Background: bg-mint-500/10
Border: border-mint-500/30
Text: text-mint-400
Icon: ● (solid circle)
Text: "Active"
```

### Booting Instance
```
Background: bg-electric-500/10
Border: border-electric-500/30
Text: text-electric-400
Icon: ◐ (half circle)
Text: "Booting"
```

### Unhealthy Instance
```
Background: bg-sunset-500/10
Border: border-sunset-500/30
Text: text-sunset-400
Icon: ⚠ (warning)
Text: "Unhealthy"
```

### Terminated Instance
```
Background: bg-gray-500/10
Border: border-gray-500/30
Text: text-gray-400
Icon: ○ (empty circle)
Text: "Terminated"
```

### Terminating Instance
```
Background: bg-red-500/10
Border: border-red-500/30
Text: text-red-400
Icon: ⏳ (hourglass)
Text: "Terminating"
```

---

## Instance Cards

### Default State
```
Background: bg-gray-900/50
Border: 1px solid border-gray-800
Cursor: cursor-pointer
Padding: p-6
Border Radius: rounded-xl
Transition: transition-all
```

### Hover State
```
Background: bg-gray-800/70
Border: border-electric-500/30
(everything else same)
```

### Selected State
```
Background: bg-sakura-500/10
Border: border-sakura-500
Additional: anime-glow class
```

---

## Dialogs/Modals

### Backdrop
```
Background: bg-black/50
Backdrop Filter: backdrop-blur-sm
Z-index: z-50
Padding: p-4
```

### Modal Container
```
Background: bg-gray-900
Border: border-electric-500/30
Border Radius: rounded-xl
Additional: anime-glow class
Max Width: max-w-2xl (launch) or max-w-3xl (config)
```

### Modal Header
```
Padding: p-6
Border Bottom: border-gray-800
```

### Modal Content
```
Padding: p-6
Space Between: space-y-6
Max Height: max-h-[70vh] (with overflow-y-auto)
```

### Modal Footer
```
Padding: p-6
Border Top: border-gray-800
Layout: flex gap-3
```

---

## Input Fields

### API Key Input
```
Type: password
Background: bg-gray-800/50
Border: border-gray-700
Border Radius: rounded-lg
Text: text-white
Padding: px-4 py-3

Focus:
  outline-none
  border-electric-500
```

### Instance Name Input (Special)
```
Background: bg-gray-900
Border: 2px solid border-electric-400/40
Text: Gradient (transparent with background-clip: text)
  from-electric-400 via-mint-400 to-electric-400
Font: font-mono text-xl font-bold
Padding: px-5 py-4

Wrapper has animated gradient border:
  from-electric-500 via-mint-400 to-sunset-500
  blur opacity-30
  group-hover:opacity-50
  animate-pulse

Focus:
  border-electric-400
  ring-2 ring-electric-500/50
```

---

## Loading States

### Page Loading
```
Layout: Centered (flex items-center justify-center)
Icon: ☁️ with animate-pulse
Text: "Loading Lambda data..."
Text Color: text-electric-400
```

### Action Loading
```
Icon: ⏳ or ⚡ (spinning)
Animation: animate-spin
Container: bg-electric-500/10 border border-electric-500/30
Progress Bar: bg-electric-500 animate-pulse width: 100%
```

### Button Loading Text
```
"⏳ Launching..."
"⏳ Terminating..."
"⏳ Restarting..."
```

---

## Feedback Messages

### Success Banner
```
Background: bg-mint-500/10
Border: border-mint-500/30
Text: text-mint-400
Icon: ✓ (checkmark)
Padding: p-4
Border Radius: rounded-lg
Auto-dismiss: 3 seconds
```

### Error Banner
```
Background: bg-sunset-500/10
Border: border-sunset-500/30
Text: text-sunset-400
Padding: p-4 or p-3
Border Radius: rounded-lg
Stays visible: User must dismiss
```

### Progress Banner (Terminating)
```
Background: bg-red-500/10
Border: border-red-500/30
Text: text-red-400
Icon: ⏳ (spinning)
Progress Bar: bg-red-500 animate-pulse
Padding: p-4
```

---

## Special UI Elements

### Hold-to-Destroy Button

**Default State:**
```
Background: bg-red-500/20
Border: 2px solid border-red-500/50
Text: text-red-400 font-bold
Text Content: "⛔ HOLD TO DESTROY"
Position: relative (for overlay)
```

**During Hold:**
```
Progress Overlay:
  position: absolute inset-0
  bg-red-500/30
  width: ${holdProgress}% (animated 0-100%)

Text: "Hold... X%" where X is progress
Z-index: relative z-10 (text stays on top)
```

**Animation:**
- Uses requestAnimationFrame for 60fps
- Duration: 2000ms (2 seconds)
- Smooth linear interpolation

### Dice Button (Name Generator)
```
Background: bg-gradient-to-r from-electric-500/20 to-mint-500/20
Border: 2px solid border-electric-500/50
Text: 🎲 (dice emoji, text-2xl)
Padding: px-5 py-4

Hover:
  from-electric-500/30 to-mint-500/30
  scale-110
  rotate-12
  transition-all
```

### Instance Name Display (in card)
```
Text Size: text-lg
Font Weight: font-semibold
Color: text-gray-200
Fallback: Shows instance.id if no name
```

### IP Address Display
```
Font: font-mono
Background: bg-gray-800
Padding: px-2 py-1
Border Radius: rounded
Color: text-mint-400 (public IP)
       text-gray-400 (private IP)
```

### SSH Command Box
```
Background: bg-electric-500/10
Border: border-electric-500/30
Padding: p-4
Border Radius: rounded-lg

Command:
  Font: font-mono text-sm
  Color: text-gray-300
  Layout: flex items-center justify-between

Copy Button:
  bg-electric-500/20
  hover:bg-electric-500/30
  text-xs text-electric-400
  Icon: 📋 Copy
```

### Jupyter Info Box
```
Background: bg-neon-500/10
Border: border-neon-500/30
Padding: p-4
Border Radius: rounded-lg

URL:
  Color: text-neon-400
  Hover: text-neon-300
  Decoration: underline
  Target: _blank

Token:
  Font: font-mono
  Truncate: text-truncate
  Copy button: bg-neon-500/20 hover:bg-neon-500/30
```

### Confirmation Dialog (First Stage)
```
Border: border-sunset-500/30
Header Theme: text-sunset-400
Icon: ⚠ in title

Warning Box:
  bg-sunset-500/10
  border: border-sunset-500/30
  padding: p-4
  font: font-mono
```

### Final Warning Dialog
```
Border: border-sunset-500/30
Overall theme: Red/danger

Warning Box:
  bg-red-500/10
  border: 2px solid border-red-500/50

Title: text-red-400 font-bold text-lg
  "⛔ FINAL WARNING"

Instance ID Container:
  bg-black/30
  border: border-red-500/30
  font-mono text-sm
```

---

## Animation Timing

### Standard Transitions
```
Duration: default (150ms)
Class: transition-all
Easing: ease-in-out
```

### Hover Effects
```
Duration: 150ms
Properties: background, border, transform
Easing: ease-in-out
```

### Modal Fade
```
Duration: 200ms
Properties: opacity, backdrop-filter
Easing: ease-out
```

### Hold-to-Destroy
```
Duration: 2000ms (2 seconds)
FPS: 60 (via requestAnimationFrame)
Easing: linear
```

### Pulse Animations
```
Properties: opacity
Duration: 2s
Iteration: infinite
Timing: cubic-bezier
```

### Spin Animations
```
Properties: transform (rotate)
Duration: 1s
Iteration: infinite
Timing: linear
```

---

## Spacing & Layout

### Page Padding
```
Main Container: p-8
Card Grid Gap: gap-4
Section Spacing: space-y-6
Button Group Gap: gap-2 or gap-3
```

### Card Internal Spacing
```
Card Padding: p-6
Section Spacing: space-y-4 (details)
                space-y-2 (compact lists)
Button Row Gap: gap-2
Flex Wrap: flex-wrap
```

### Grid Layouts
```
Instance Cards: grid-cols-1 lg:grid-cols-2
Config Details: grid-cols-2 gap-4
```

---

## Border Radius

### Standard Elements
```
Cards: rounded-xl (0.75rem / 12px)
Buttons: rounded-lg (0.5rem / 8px)
Badges: rounded-full (9999px)
Inputs: rounded-lg (0.5rem / 8px)
```

---

## Shadows & Effects

### Glow Effect (anime-glow class)
```
Expected appearance:
  Subtle colored shadow around element
  Matches border color theme
  Increases on hover (some elements)
```

### Backdrop Blur
```
Amount: backdrop-blur-sm
Applied to: Dialog backdrops, header, footer
Effect: Blurs background behind element
```

### Progress Bar Shine
```
Effect: animate-pulse on background
Creates: Pulsing brightness effect
```

---

## Responsive Breakpoints

### Tailwind Defaults
```
sm: 640px
md: 768px
lg: 1024px
xl: 1280px
2xl: 1536px
```

### Used in Component
```
lg:grid-cols-2 (instance cards at 1024px+)
```

---

## Icons & Emojis

### Feature Icons
- Cloud: ☁️ (Lambda Cloud branding)
- Rocket: 🚀 (Empty state)
- Gear: ⚙️ (Configuration)
- Clipboard: 📋 (Copy actions)
- Monitor: 📊 (Server monitoring)
- Key: 🔑 (SSH section)
- Lock: 🔐 (Individual SSH key)
- Notebook: 📓 (Jupyter)
- Hardware: 🖥️ (Hardware section)
- Document: 📋 (Instance details)

### Status Icons
- Active: ● (solid bullet)
- Booting: ◐ (half circle)
- Unhealthy: ⚠ (warning triangle)
- Terminated: ○ (empty circle)
- Terminating: ⏳ (hourglass)

### Action Icons
- Warning: ⚠ (terminate action)
- Refresh: 🔄 (restart action)
- Stop Sign: ⛔ (hold to destroy)
- Hourglass: ⏳ (loading state)
- Lightning: ⚡ (launching)
- Checkmark: ✓ (success)
- Dice: 🎲 (random name)

---

## Z-Index Layers

```
Base layer: z-0 (instance cards, content)
Elevated: z-10 (dropdowns, tooltips)
Modal backdrop: z-50 (dialog backgrounds)
Modal content: above backdrop (dialog content)
Progress overlay: relative z-10 (button progress text)
```

---

## Common Patterns

### Section Header
```
<h3 class="text-lg font-semibold text-gray-200 flex items-center gap-2">
  <span>{EMOJI}</span>
  {SECTION_TITLE}
</h3>
```

### Data Display Box
```
<div class="p-4 bg-gray-800/50 rounded-lg">
  <div class="text-xs text-gray-500 mb-1">{LABEL}</div>
  <div class="text-sm text-gray-300">{VALUE}</div>
</div>
```

### Button with Loading State
```
<button disabled={loading}>
  {loading ? '⏳ Loading...' : 'Action Text'}
</button>
```

### Copy Button Pattern
```
<button onClick={() => navigator.clipboard.writeText(text)}>
  📋 Copy
</button>
```

---

## Print Reference

When testing, refer to this guide for expected appearance:

✓ **All buttons** should have visible hover effects
✓ **All status badges** should have the correct color for their state
✓ **All transitions** should be smooth (not janky)
✓ **All loading states** should show animated spinners
✓ **All text** should be readable (good contrast)
✓ **All interactive elements** should change cursor to pointer
✓ **All disabled elements** should show opacity-50 and cursor-not-allowed

---

**End of Visual Reference Guide**

*Use this alongside VISUAL_TESTING_CHECKLIST.md for comprehensive UI testing*
