<script>
  import {createEventDispatcher} from "svelte"
  const dispatch = createEventDispatcher()

  export let angle = 270; // your script goes here
  export let minAngle = 0;
  export let maxAngle = 270;
  export let offset = 45;

  export let width = 200;
  export let height = 200;

  export let knobRadius = 70;
  export let knobInset = 20;
  export let arcRadius = 90;
  export let indicatorRadius = 10;
  export let volumeThickness = 5;

  export let knobColor = "#000";
  export let indicatorColor = "#fff";
  export let volumeColor = "#000";
  export let textColor = "#fff";

  export let value = 0;
  export let minValue = -10
  export let maxValue = 0

  let arcLength = 2 * Math.PI * arcRadius;
  let arcPart = (angle / 360) * arcLength;

  const checkOutOfBounds = () => {
    if (angle > maxAngle) {
      angle = maxAngle;
    } else if (angle < minAngle) {
      angle = minAngle;
    }
  }
 
  const handleWheel = e => {
    if (e.deltaY > 0) {
      angle = angle - 5;
    } else if (e.deltaY < 0) {
      angle = angle + 5;
    }

    checkOutOfBounds()
    

value = map_range(angle, minAngle, maxAngle, minValue, maxValue)
    dispatch("valueChanged")
  };

  const handleTouchmove = (e) => {
      const xCoord = e.touches[0].clientX
      const screenHalf = screen.width / 2
      if (xCoord > screenHalf) {
          angle = angle + 3
      } else if (xCoord < screenHalf) {
          angle = angle - 3
      }

      checkOutOfBounds()

    value = map_range(angle, minAngle, maxAngle, minValue, maxValue)
    dispatch("valueChanged")
  }

  function map_range(value, low1, high1, low2, high2) {
    return low2 + ((high2 - low2) * (value - low1)) / (high1 - low1);
  }

  $: angle = map_range(value, minValue,maxValue,minAngle,maxAngle);
</script>

<svg
  {height}
  {width}
  on:wheel|preventDefault={handleWheel}
  on:touchmove|preventDefault={handleTouchmove}
  viewBox={width / -2 + ' ' + height / -2 + ' ' + width + ' ' + height}>
  <circle

    cx="0"
    cy="0"
    r={knobRadius}
    fill={knobColor} />
  <text text-anchor="middle" fill={textColor}>
    {Math.round(value*100)/100}
  </text>
  <circle
    cx={(knobRadius - knobInset) * -1}
    cy="0"
    r={indicatorRadius}
    fill={indicatorColor}
    transform="rotate({angle - offset})" />
  <circle
    cx="0"
    cy="0"
    r={arcRadius}
    fill="none"
    stroke={volumeColor}
    transform="rotate({180 - offset})"
    stroke-width={volumeThickness}
    stroke-linecap="round"
    stroke-dasharray={arcLength}
    stroke-dashoffset={(arcLength * -angle) / 360 - arcLength} />

</svg>
