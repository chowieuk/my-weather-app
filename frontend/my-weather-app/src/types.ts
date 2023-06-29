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
    moon_phase: MoonPhase;
    moon_illumination: number;
};

export type MoonPhase =
    | "New Moon"
    | "Waxing Crescent"
    | "First Quarter"
    | "Waxing Gibbous"
    | "Full Moon"
    | "Waning Gibbous"
    | "Last Quarter"
    | "Waning Crescent";
