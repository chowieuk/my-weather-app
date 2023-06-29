export type AstroData = {
    name: string;
    region: string;
    country: string;
    date: string;
    astro: {
        sunrise: string;
        sunset: string;
        moonrise: string;
        moonset: string;
        moon_phase: string;
        moon_illumination: number;
    };
    // expires_at: string;
};
