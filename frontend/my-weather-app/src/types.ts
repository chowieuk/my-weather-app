export type AstroData = {
    name: string;
    region: string;
    country: string;
    date: string;
    astro: AstroDetails;
    // expires_at: string;
};

export type AstroDetails = {
    sunrise: string;
    sunset: string;
    moonrise: string;
    moonset: string;
    moon_phase: string;
    moon_illumination: number;
};
