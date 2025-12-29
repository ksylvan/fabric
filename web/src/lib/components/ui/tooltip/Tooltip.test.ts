import { describe, expect, it } from 'vitest';
import { calculateTooltipPosition, formatPositionStyle, TOOLTIP_GAP } from './positioning';

describe('Tooltip positioning logic', () => {
	const mockRect = {
		top: 100,
		bottom: 130,
		left: 200,
		right: 300,
		width: 100,
		height: 30,
		x: 200,
		y: 100,
		toJSON: () => ({})
	} as DOMRect;

	it('calculates top position correctly', () => {
		const result = calculateTooltipPosition(mockRect, 'top');
		expect(result.top).toBe(92); // 100 - 8
		expect(result.left).toBe(250); // 200 + 100/2
	});

	it('calculates bottom position correctly', () => {
		const result = calculateTooltipPosition(mockRect, 'bottom');
		expect(result.top).toBe(138); // 130 + 8
		expect(result.left).toBe(250); // 200 + 100/2
	});

	it('calculates left position correctly', () => {
		const result = calculateTooltipPosition(mockRect, 'left');
		expect(result.top).toBe(115); // 100 + 30/2
		expect(result.left).toBe(192); // 200 - 8
	});

	it('calculates right position correctly', () => {
		const result = calculateTooltipPosition(mockRect, 'right');
		expect(result.top).toBe(115); // 100 + 30/2
		expect(result.left).toBe(308); // 300 + 8
	});

	it('uses the correct gap value', () => {
		expect(TOOLTIP_GAP).toBe(8);
	});

	it('formats position style correctly', () => {
		const position = { top: 100, left: 200 };
		const style = formatPositionStyle(position);
		expect(style).toBe('top: 100px; left: 200px;');
	});

	it('respects custom gap parameter', () => {
		const customGap = 16;
		const result = calculateTooltipPosition(mockRect, 'top', customGap);
		expect(result.top).toBe(84); // 100 - 16
	});
});
