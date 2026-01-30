# Lambda Cloud Visual Testing Checklist

**Purpose:** Quick reference checklist for manual UI testing
**Estimated Time:** 45 minutes
**Date:** 2025-11-20

---

## Pre-Testing Setup

- [ ] Application builds successfully
- [ ] No console errors on startup
- [ ] Lambda API key available for testing

---

## 1. API Key Management (5 min)

### Initial Dialog
- [ ] **Dialog appears** automatically if no API key set
- [ ] **Cloud emoji** (☁️) displays correctly
- [ ] **Input field** accepts text
- [ ] **Password masking** works (text hidden)
- [ ] **Border color** changes on focus (electric blue)
- [ ] **Button disabled** when input empty (opacity 50%, cursor not-allowed)
- [ ] **Button enabled** when text entered
- [ ] **Hover effect** works on button (background lightens)

### Validation
- [ ] **Enter key** submits form
- [ ] **Invalid key** shows error message in sunset-colored box
- [ ] **Valid key** closes dialog
- [ ] **Data loads** after successful connection
- [ ] **Loading state** visible during data fetch

### Visual Elements
- [ ] Electric blue color theme consistent
- [ ] Border glow animation (`anime-glow`) visible
- [ ] Smooth transitions on all state changes

**Notes:**
- Missing: Loading spinner during API key validation

---

## 2. Instance Listing (5 min)

### Main View
- [ ] **Header displays** with cloud emoji and title
- [ ] **Instance count** shown in subtitle
- [ ] **Refresh button** visible (🔄)
- [ ] **Launch button** visible (+ Launch Instance)

### Empty State
- [ ] **Rocket emoji** (🚀) displays if no instances
- [ ] **Message** "No instances running" visible
- [ ] **CTA button** "Launch Instance" works
- [ ] **Centered layout** looks good

### Instance Cards (if instances exist)
- [ ] **Cards display** in grid (1 col mobile, 2 col desktop)
- [ ] **Instance name** or ID shown
- [ ] **Instance type** shown below name
- [ ] **Status badge** displayed with icon and color
- [ ] **IP address** shown in mint green monospace
- [ ] **Region** displayed
- [ ] **SSH keys** listed

### Status Colors
- [ ] **Active:** Mint green (●)
- [ ] **Booting:** Electric blue (◐)
- [ ] **Unhealthy:** Sunset orange (⚠)
- [ ] **Terminated:** Gray (○)
- [ ] **Terminating:** Red (⏳)

### Auto-Refresh
- [ ] **Wait 10 seconds** after initial load
- [ ] **Status updates** without manual refresh
- [ ] **No error flash** if transient network issue
- [ ] **Smooth state transitions**

**Notes:**
- Missing: "Last updated" timestamp indicator
- Missing: Visual indicator that auto-refresh is active

---

## 3. Launch Instance Dialog (10 min)

### Opening Dialog
- [ ] **Click launch button** in header
- [ ] **Dialog opens** with smooth animation
- [ ] **Backdrop blur** visible
- [ ] **Modal centered** on screen
- [ ] **Electric blue border** with glow effect

### Instance Type Selection
- [ ] **Types load** and display
- [ ] **First type auto-selected** (electric blue highlight)
- [ ] **Selection ring** visible on selected type (2px ring)
- [ ] **Hover effect** on unselected types (border changes)
- [ ] **Price shown** in mint green ($/hr format)
- [ ] **Specs displayed** (vCPUs, RAM, Storage, GPUs)
- [ ] **Available regions** listed below specs
- [ ] **Scroll works** if many types available
- [ ] **Priority order:** GH200 > H100 > B200 > A100 > others

### Click Different Types
- [ ] **Selection changes** on click
- [ ] **Region auto-updates** for new type
- [ ] **Visual feedback** immediate
- [ ] **Smooth transitions**

### Region Selection
- [ ] **Regions display** for selected type
- [ ] **First region auto-selected** (thicker blue border)
- [ ] **Full width** if only 1 region available
- [ ] **Hover effect** on unselected regions
- [ ] **Click changes** selection
- [ ] **Border thickness** increases on selection (2px)

### SSH Key Selection
- [ ] **Keys display** as toggle buttons
- [ ] **First key auto-selected**
- [ ] **Click toggles** selection on/off
- [ ] **Multiple keys** can be selected
- [ ] **Selection shows** thicker border (2px)
- [ ] **Hover effect** works on all keys
- [ ] **Flex wrap** layout looks good

### Instance Name
- [ ] **Random name generated** on open (three words)
- [ ] **Gradient text effect** visible
- [ ] **Animated gradient border** pulses
- [ ] **Focus ring** appears when clicked
- [ ] **User can edit** name
- [ ] **Dice button** (🎲) visible
- [ ] **Dice rotates** on hover (scale + rotate)
- [ ] **Click dice** generates new random name

### Launch Button
- [ ] **Button disabled** if requirements not met
- [ ] **Disabled state** shows opacity 50%
- [ ] **Enabled** when type + keys selected
- [ ] **Hover effect** works when enabled
- [ ] **Click launches** instance

### Loading State
- [ ] **Button text** changes to "⏳ Launching..."
- [ ] **Loading box appears** with electric blue theme
- [ ] **Lightning bolt** spins (⚡ animate-spin)
- [ ] **Progress bar** animates (pulse effect)
- [ ] **Button disabled** during launch

### Success State
- [ ] **Success message** appears in mint green box
- [ ] **Checkmark** in message (✓)
- [ ] **Button text** changes to "✓ Success!"
- [ ] **Dialog auto-closes** after 2 seconds
- [ ] **Instance list refreshes**

### Error State
- [ ] **Error message** appears in sunset box
- [ ] **Message text** clearly describes error
- [ ] **Dialog stays open**
- [ ] **User can retry**

### Summary Panel
- [ ] **Panel shows** selected instance type
- [ ] **Cost per hour** displayed in mint green
- [ ] **SSH keys count** shown
- [ ] **Updates dynamically** when selections change

### Dialog Close
- [ ] **Cancel button** closes dialog
- [ ] **Cancel has** hover effect
- [ ] **Backdrop click** doesn't close during launch
- [ ] **X button** would close (not present in launch dialog)

**Notes:**
- Instance name input is exceptionally polished
- Hold-to-destroy pattern should be referenced for destructive actions

---

## 4. Terminate Instance Flow (10 min)

### Initial Terminate Button
- [ ] **Button visible** on all instance cards
- [ ] **Sunset theme** (orange colors)
- [ ] **Warning icon** (⚠) shown
- [ ] **Hover effect** increases background opacity
- [ ] **Disabled** during termination (opacity 50%)
- [ ] **Loading text** "⏳ Terminating..." when active
- [ ] **Click stops** event propagation (card not clicked)

### First Confirmation Dialog
- [ ] **Dialog opens** with blur backdrop
- [ ] **Sunset border theme** visible
- [ ] **Warning emoji** in header (⚠)
- [ ] **Clear warning text** about data loss
- [ ] **Instance ID** shown in sunset box
- [ ] **Monospace font** for ID
- [ ] **Two buttons:** Cancel and Continue

### Cancel Button
- [ ] **Gray theme**
- [ ] **Hover effect** works
- [ ] **Click closes** dialog

### Continue Button
- [ ] **Sunset theme**
- [ ] **Hover effect** works
- [ ] **Click advances** to final confirmation

### Final Confirmation Dialog
- [ ] **Red theme** applied (more intense)
- [ ] **"FINAL WARNING" text** bold and prominent
- [ ] **Double border** visible (border-2)
- [ ] **Stronger warning text**
- [ ] **Instance ID** in black box with red border
- [ ] **Hold instruction** clearly stated

### Hold-to-Destroy Button
- [ ] **Button text:** "⛔ HOLD TO DESTROY"
- [ ] **Red theme** with bold text
- [ ] **Mouse down** starts progress
- [ ] **Progress bar** fills from left to right
- [ ] **Progress overlay** visible (red 30% opacity)
- [ ] **Button text updates** to "Hold... X%"
- [ ] **Percentage increases** smoothly (0-100%)
- [ ] **60 FPS animation** (smooth, not janky)

### Early Release
- [ ] **Mouse up** cancels progress
- [ ] **Mouse leave** cancels progress
- [ ] **Progress resets** to 0%
- [ ] **Button returns** to "HOLD TO DESTROY"

### Complete Hold
- [ ] **Progress reaches** 100%
- [ ] **Termination triggers** automatically
- [ ] **Button text** changes to "⏳ Terminating..."
- [ ] **Button disabled**

### Touch Support
- [ ] **Touch start** works same as mouse down
- [ ] **Touch end** works same as mouse up
- [ ] **Touch drag away** cancels like mouse leave

### Termination Progress
- [ ] **Progress banner appears** (red theme)
- [ ] **Spinning hourglass** visible (⏳)
- [ ] **Progress text** shown
- [ ] **Progress bar** animates
- [ ] **Dialog closes**

### Instance Status Update
- [ ] **Status badge** changes to "terminating"
- [ ] **Red theme** applied to badge
- [ ] **Hourglass icon** shown
- [ ] **Instance still** in list

### Success Message
- [ ] **Success banner** appears (mint green)
- [ ] **Checkmark** in message (✓)
- [ ] **Instance ID** included
- [ ] **Auto-dismisses** after 3 seconds

### Error Handling
- [ ] **Error banner** appears if API fails (sunset)
- [ ] **Error message** clearly describes issue
- [ ] **Dialog closes** anyway
- [ ] **Instance status** doesn't change

**Notes:**
- Hold-to-destroy is best-in-class UX for destructive action
- Smooth animation using requestAnimationFrame
- Excellent multi-stage confirmation flow

---

## 5. Restart Instance (5 min)

### Restart Button
- [ ] **Button visible** on all instance cards
- [ ] **Gray theme**
- [ ] **Restart icon** (🔄)
- [ ] **Hover effect** lightens background
- [ ] **Disabled** when instance not active
- [ ] **Disabled** during restart operation
- [ ] **Cursor not-allowed** when disabled

### Click Restart
- [ ] **Button text** changes to "⏳ Restarting..."
- [ ] **Button disabled**
- [ ] **No dialog** appears (direct action)

### During Restart
- [ ] **Loading state** maintained
- [ ] **Other instances** still interactive

### Success
- [ ] **Success banner** appears (mint green)
- [ ] **Checkmark** visible (✓)
- [ ] **Instance ID** in message
- [ ] **Auto-dismisses** after 3 seconds
- [ ] **Instance list** refreshes
- [ ] **Button returns** to normal state

### Error
- [ ] **Error banner** appears (sunset)
- [ ] **Error message** clearly describes issue
- [ ] **Button returns** to normal state

**Notes:**
- Simple, effective feedback
- No confirmation needed (restart is non-destructive)

---

## 6. Instance Configuration Modal (5 min)

### Opening Modal
- [ ] **Click anywhere** on instance card
- [ ] **Action buttons** don't trigger modal (stopPropagation works)
- [ ] **Modal opens** with smooth animation
- [ ] **Backdrop blur** visible
- [ ] **Modal centered**
- [ ] **Electric blue border** with glow

### Modal Header
- [ ] **Gear emoji** (⚙️) visible
- [ ] **Title** "Instance Configuration"
- [ ] **Hostname or ID** shown as subtitle
- [ ] **X button** in top right
- [ ] **X button hover** shows gray background

### Instance Details Section
- [ ] **Section header** with emoji (📋)
- [ ] **2x2 grid layout**
- [ ] **Instance ID** displayed
- [ ] **Status** capitalized
- [ ] **Public IP** in mint green monospace
- [ ] **Private IP** in gray monospace
- [ ] **"N/A" shown** for missing data

### Hardware Section
- [ ] **Section header** with emoji (🖥️)
- [ ] **2x2 grid layout**
- [ ] **Instance type** name and description
- [ ] **Region** name and description
- [ ] **All text** readable

### SSH Access Section
- [ ] **Section header** with emoji (🔑)
- [ ] **SSH keys** listed with lock icon (🔐)
- [ ] **"No SSH keys"** message if empty
- [ ] **SSH command box** appears if IP exists
- [ ] **Command** in monospace: `ssh ubuntu@{ip}`
- [ ] **Copy button** visible (📋 Copy)
- [ ] **Copy button hover** effect works
- [ ] **Click copies** to clipboard

### Jupyter Section (if available)
- [ ] **Section appears** only if jupyter_url exists
- [ ] **Section header** with emoji (📓)
- [ ] **Neon color theme** (different from electric)
- [ ] **URL clickable** with underline
- [ ] **URL opens** in new tab
- [ ] **Token displayed** in monospace
- [ ] **Token truncated** if too long
- [ ] **Copy button** visible
- [ ] **Copy button hover** works
- [ ] **Click copies** token

### Modal Footer
- [ ] **Close button** visible
- [ ] **Full width** button
- [ ] **Hover effect** works
- [ ] **Click closes** modal

### Closing Modal
- [ ] **Close button** works
- [ ] **X button** works
- [ ] **Escape key** closes modal
- [ ] **Backdrop click** closes modal
- [ ] **Clicking inside** modal doesn't close
- [ ] **Smooth fade-out** animation

### Content Scrolling
- [ ] **Content scrolls** if needed (max-h-70vh)
- [ ] **Scroll visible** if content long
- [ ] **Footer stays** at bottom

**Notes:**
- Missing: Copy confirmation feedback
- Suggestion: Toast message "Copied!" after clipboard operations

---

## 7. Hover Effects & Polish (5 min)

### Instance Cards
- [ ] **Hover lightens** background
- [ ] **Border changes** to electric blue
- [ ] **Cursor** changes to pointer
- [ ] **Transition** is smooth
- [ ] **Selected cards** have sakura theme
- [ ] **Selected cards** have glow effect
- [ ] **Glow animation** visible

### All Buttons
- [ ] **Primary buttons** (electric blue) lighten on hover
- [ ] **Secondary buttons** (gray) lighten on hover
- [ ] **Destructive buttons** (sunset) lighten on hover
- [ ] **Copy buttons** (mint) lighten on hover
- [ ] **Selection buttons** show border change
- [ ] **Disabled buttons** show not-allowed cursor
- [ ] **Dice button** scales and rotates

### Input Fields
- [ ] **API key input** border changes on focus
- [ ] **Instance name input** shows ring on focus
- [ ] **Instance name** gradient border animates
- [ ] **Smooth transitions** on all inputs

### Instance Type Cards
- [ ] **Unselected hover** shows border change
- [ ] **Selected shows** ring effect
- [ ] **Click feedback** immediate
- [ ] **Transitions smooth**

### Region Buttons
- [ ] **Hover** changes border color
- [ ] **Selected** has thicker border
- [ ] **Transitions smooth**

### SSH Key Buttons
- [ ] **Hover** changes border color
- [ ] **Selected** has thicker border
- [ ] **Multiple selections** visually clear
- [ ] **Transitions smooth**

### Loading Animations
- [ ] **Cloud emoji** pulses on initial load
- [ ] **Lightning bolt** spins smoothly
- [ ] **Hourglass** spins smoothly
- [ ] **Progress bars** animate smoothly
- [ ] **Gradient border** pulses on name input

### Glow Effects
- [ ] **anime-glow class** creates subtle glow
- [ ] **Glow visible** on primary buttons
- [ ] **Glow visible** on dialogs
- [ ] **Glow visible** on selected cards

**Notes:**
- Polish level is exceptional
- Consistent animation timing across all elements
- Smooth 60fps animations

---

## 8. Additional UI Elements (5 min)

### Header
- [ ] **Cloud emoji** (☁️) in title
- [ ] **"Lambda Cloud" title** in electric blue
- [ ] **Subtitle** "Manage your GPU instances"
- [ ] **Refresh button** works
- [ ] **Launch button** has glow effect

### Success/Error Banners
- [ ] **Success** uses mint green theme
- [ ] **Error** uses sunset orange theme
- [ ] **Progress** uses red theme
- [ ] **Auto-dismiss** after 3 seconds (success)
- [ ] **Smooth fade** in and out

### Monitor Button
- [ ] **Only visible** for active instances with IP
- [ ] **Electric blue** theme with glow
- [ ] **Emoji** (📊) visible
- [ ] **Hover effect** works
- [ ] **Click opens** ServerMonitor component

### Copy SSH Button (on card)
- [ ] **Only visible** for active instances with IP
- [ ] **Mint green** theme
- [ ] **Emoji** (📋) visible
- [ ] **Hover effect** works
- [ ] **Click copies** SSH command

### Select for Packages Button
- [ ] **Only visible** for active instances
- [ ] **Full width** on card
- [ ] **Unselected** shows gray with hover effect
- [ ] **Selected** shows sakura theme with glow
- [ ] **Checkmark** appears when selected
- [ ] **Text changes** appropriately

---

## 9. Responsive Design (3 min)

### Desktop (>1024px)
- [ ] **Instance grid** shows 2 columns
- [ ] **All buttons** visible and properly sized
- [ ] **Dialogs** properly sized
- [ ] **Text** readable

### Tablet (768-1024px)
- [ ] **Instance grid** shows 1 column
- [ ] **Layout** still looks good
- [ ] **Buttons** properly sized

### Mobile (<768px)
- [ ] **Instance grid** shows 1 column
- [ ] **Buttons** wrap correctly
- [ ] **Dialogs** fill most of screen
- [ ] **Content scrollable**
- [ ] **Text** remains readable

### Dialog Scrolling
- [ ] **Launch dialog** scrolls if content tall
- [ ] **Config modal** scrolls if content tall
- [ ] **Footer buttons** stay at bottom

---

## 10. Edge Cases (5 min)

### No Instances Available
- [ ] **Empty state** shows
- [ ] **Launch button** still works
- [ ] **No errors** in console

### No Instance Types Available
- [ ] **Warning shown** in launch dialog
- [ ] **Launch button** disabled
- [ ] **Clear message** to user

### API Errors
- [ ] **Error banners** appear
- [ ] **Error text** clearly describes issue
- [ ] **User can** retry actions
- [ ] **No crashes**

### Network Disconnection
- [ ] **Auto-refresh** continues to retry silently
- [ ] **No error spam**
- [ ] **Manual refresh** shows error if offline

### Multiple Quick Clicks
- [ ] **Launch prevents** duplicate submissions
- [ ] **Terminate** can't be triggered multiple times
- [ ] **Restart** can't be triggered multiple times

### Very Long Names
- [ ] **Instance names** don't break layout
- [ ] **Instance IDs** don't overflow
- [ ] **Text truncates** or wraps appropriately

---

## 11. Console & Debug (2 min)

### Console Logs
- [ ] **No errors** on normal operations
- [ ] **No warnings** about React keys or props
- [ ] **Debug logs** present but not excessive
- [ ] **API responses** logged in development

### Network Tab
- [ ] **API calls** go to correct endpoints
- [ ] **Auth headers** included
- [ ] **Responses** are JSON
- [ ] **Polling** happens every 10 seconds

---

## Post-Testing

### Summary
- [ ] All critical features working
- [ ] All visual feedback present
- [ ] No critical bugs found
- [ ] Performance acceptable
- [ ] Ready for deployment

### Issues Log
Document any issues found:

1. Issue:
   - Severity: Critical / High / Medium / Low
   - Description:
   - Steps to reproduce:

2. Issue:
   - Severity:
   - Description:
   - Steps to reproduce:

### Improvements Needed
List any enhancements identified:

1. Enhancement:
   - Priority: High / Medium / Low
   - Description:
   - Rationale:

---

## Sign-Off

**Tester Name:** _______________________
**Date:** _______________________
**Result:** ⬜ Pass  ⬜ Pass with Minor Issues  ⬜ Fail
**Notes:**

---

**End of Checklist**

*Estimated completion time: 45 minutes*
*For issues found, refer to LAMBDA_FEATURES_TEST_REPORT.md for detailed analysis*
