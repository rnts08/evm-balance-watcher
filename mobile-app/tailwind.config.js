/** @type {import('tailwindcss').Config} */
module.exports = {
    // NOTE: Update this to include the paths to all of your component files.
    content: ["./App.{js,jsx,ts,tsx}", "./app/**/*.{js,jsx,ts,tsx}", "./src/**/*.{js,jsx,ts,tsx}"],
    presets: [require("nativewind/preset")],
    theme: {
        extend: {
            colors: {
                background: "#09090b",
                foreground: "#fafafa",
                card: "#18181b",
                "card-foreground": "#fafafa",
                primary: "#3b82f6",
                "primary-foreground": "#ffffff",
                secondary: "#27272a",
                border: "#27272a",
                accent: "#a855f7",
                success: "#22c55e",
                error: "#ef4444",
            },
        },
    },
    plugins: [],
};
