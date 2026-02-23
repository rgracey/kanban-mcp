const KEY = 'darkMode';

function init(): boolean {
	if (typeof localStorage === 'undefined') return false;
	const stored = localStorage.getItem(KEY);
	return stored !== null
		? stored === 'true'
		: window.matchMedia('(prefers-color-scheme: dark)').matches;
}

function apply(dark: boolean) {
	document.documentElement.classList.toggle('dark', dark);
	localStorage.setItem(KEY, String(dark));
}

export function initDarkMode() {
	const dark = init();
	apply(dark);
	return dark;
}

export function toggleDarkMode(current: boolean): boolean {
	const next = !current;
	apply(next);
	return next;
}
