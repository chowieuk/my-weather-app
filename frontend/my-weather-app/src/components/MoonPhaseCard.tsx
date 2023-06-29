import React from "react";

import NewMoon from "../assets/phases/new-moon.svg";
import WaxingCrescent from "../assets/phases/waxing-crescent.svg";
import FirstQuarter from "../assets/phases/first-quarter.svg";
import WaxingGibbous from "../assets/phases/waxing-gibbous.svg";
import FullMoon from "../assets/phases/full-moon.svg";
import WaningGibbous from "../assets/phases/waning-gibbous.svg";
import LastQuarter from "../assets/phases/last-quarter.svg";
import WaningCrescent from "../assets/phases/waning-crescent.svg";

import { MoonPhase } from "../types";

type MoonPhaseImageMap = {
    [phase in MoonPhase]: string;
};

const moonPhaseImageMap: MoonPhaseImageMap = {
    "New Moon": NewMoon,
    "Waxing Crescent": WaxingCrescent,
    "First Quarter": FirstQuarter,
    "Waxing Gibbous": WaxingGibbous,
    "Full Moon": FullMoon,
    "Waning Gibbous": WaningGibbous,
    "Last Quarter": LastQuarter,
    "Waning Crescent": WaningCrescent,
};

type MoonPhaseCardProps = {
    phase: MoonPhase;
};

export const MoonPhaseCard: React.FC<MoonPhaseCardProps> = ({ phase }) => (
    <div>
        <img src={moonPhaseImageMap[phase]} alt={phase} />
        <h3>Moon Phase:</h3>
        <p>{phase}</p>
    </div>
);
