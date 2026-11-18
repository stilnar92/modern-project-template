import { expect, test } from "vitest";
import { m } from "./index";

// Russian has three plural forms: one, few, many (frontend standard 28).
test.each([
  [1, "1 фотография"],
  [21, "21 фотография"],
  [2, "2 фотографии"],
  [24, "24 фотографии"],
  [5, "5 фотографий"],
  [11, "11 фотографий"],
  [12, "12 фотографий"],
  [100, "100 фотографий"],
])("%i → %s", (count, text) => {
  expect(m.photos_count({ count })).toBe(text);
});

test("a plain message", () => {
  expect(m.core_status_ok()).toBe("Ядро: ok");
});
