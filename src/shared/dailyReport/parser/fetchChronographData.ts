import { chromium } from 'playwright';
import { TimeSpentChronograph } from '../types';

const selectors = {
  counter: '.counter-header',
};

const requiredHeaderValuesLength = 3;
const requiredElementsLength = 2;

export async function fetchChronographData(
  url: string
): Promise<TimeSpentChronograph | null> {
  const browser = await chromium.launch({
    headless: true,
  });

  try {
    const ctx = await browser.newContext({
      viewport: {
        height: 800,
        width: 600,
      },
    });

    const page = await ctx.newPage();

    await page.goto(url, {
      waitUntil: 'networkidle',
    });

    const elements = await page.evaluate(selector => {
      return Array.from(document.querySelectorAll(selector)).map(
        element => element.textContent
      );
    }, selectors.counter);

    const elementTexts = elements.filter(v => v) as string[];

    if (elementTexts.length !== requiredElementsLength) {
      return null;
    }

    const [net, gross] = elementTexts.map(text => parseHeader(text));

    if (!net || !gross) {
      return null;
    }

    return {
      netSeconds: net,
      grossSeconds: gross,
    };
  } finally {
    await browser.close();
  }
}

function parseHeader(headerContent: string): number | null {
  const values = headerContent.split(/[A-z]/).filter(value => value);

  if (values.length !== requiredHeaderValuesLength) {
    return null;
  }

  const [hours, minutes, seconds] = values.map(v => parseFloat(v));

  return hours * 3600 + minutes * 60 + seconds;
}
