// PragmaOS modal focus trap (no Alpine.js trap plugin dependency).
// Used by the modal component template via x-data="modalTrap()".
function modalTrap() {
  return {
    open: false,
    focusables: [],
    focusFirst() {
      this.focusables = this.getFocusables();
      if (this.focusables.length > 0) {
        this.focusables[0].focus();
      }
    },
    getFocusables() {
      var dialog = this.$refs.dialog;
      if (!dialog) return [];
      var sel = 'button, [href], input, select, textarea, [tabindex]';
      var els = Array.from(dialog.querySelectorAll(sel));
      return els.filter(function (el) {
        return !el.disabled && el.tabindex !== '-1';
      });
    },
    cycleFocus(e) {
      this.focusables = this.getFocusables();
      if (this.focusables.length === 0) return;
      var first = this.focusables[0];
      var last = this.focusables[this.focusables.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        first.focus();
      }
    }
  };
}
