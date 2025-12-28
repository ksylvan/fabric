export const TOOLTIP_GAP = 8;

export type TooltipPosition = 'top' | 'bottom' | 'left' | 'right';

export interface Position {
	top: number;
	left: number;
}

export function calculateTooltipPosition(
	rect: DOMRect,
	position: TooltipPosition,
	gap: number = TOOLTIP_GAP
): Position {
	const positions: Record<TooltipPosition, Position> = {
		top: { top: rect.top - gap, left: rect.left + rect.width / 2 },
		bottom: { top: rect.bottom + gap, left: rect.left + rect.width / 2 },
		left: { top: rect.top + rect.height / 2, left: rect.left - gap },
		right: { top: rect.top + rect.height / 2, left: rect.right + gap }
	};

	return positions[position];
}

export function formatPositionStyle(position: Position): string {
	return `top: ${position.top}px; left: ${position.left}px;`;
}
