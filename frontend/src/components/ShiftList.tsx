import { mockShifts } from "@/data/mock-shifts";
import { ShiftCard } from "./ShiftCard";

export function ShiftList() {
  return (
    <ul className="grid list-none grid-cols-1 gap-4 md:grid-cols-2 md:gap-5 lg:grid-cols-3">
      {mockShifts.map((shift) => (
        <li key={shift.id}>
          <ShiftCard shift={shift} />
        </li>
      ))}
    </ul>
  );
}
