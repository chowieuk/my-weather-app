import { AstroData } from "./types";
import { CONFIG } from "./config";

export const getAstroData = async (location: string): Promise<AstroData> => {
    const response = await fetch(
        `${CONFIG.API_ENDPOINT}/astro?location=${location}`
    );
    const data: AstroData = await response.json();
    return data;
};
