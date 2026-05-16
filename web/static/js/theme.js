// Theme toggle: flips between light and dark, persists to localStorage.
// The initial theme is applied by an inline <head> script in layout.html.tmpl
// before CSS loads (to prevent flash of wrong theme). Icon swap is pure CSS,
// driven off the [data-theme] attribute on <html>.

function toggleTheme() {
    var current = document.documentElement.getAttribute('data-theme') || 'light';
    var next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);
}
