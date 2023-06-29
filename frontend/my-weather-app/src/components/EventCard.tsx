import React from "react";

type EventCardProps = {
    eventName: string;
    eventTime: string;
    image: string; // Path to the image file
};

export const EventCard: React.FC<EventCardProps> = ({
    eventName,
    eventTime,
    image,
}) => (
    <div>
        <img src={image} alt={eventName} />
        <h3>{eventName}</h3>
        <p>{eventTime}</p>
    </div>
);
