import React, { useState, useEffect } from 'react'
import * as d3 from 'd3'
import FlameGraph from './flamegraph'
import "../css/overview.scss"

const GROUP = 'memory'
const ROUTINE = 'memleak_profile'

const Axis = ({ orient, translate, scale, cssClass, tickFormat }) => {
  let axisElement
  const renderAxis = () => {
    let axisType = `axis${orient}`

    d3.select(axisElement).call(d3[axisType](scale).tickFormat(tickFormat))
  }
  const [state, setState] = useState()
  useEffect(() => {
    renderAxis()
  }, [])
  useEffect(() => {
    renderAxis()
  }, [state])

  return <g className={cssClass} ref={el => axisElement = el} transform={translate} />
}

const XYAxisChart = ({ scales, margins, dimensions, data }) => {
  const xAxisProps = {
    orient: 'Bottom',
    translate: `translate(0, ${dimensions.height - margins.bottom})`,
    scale: scales.xScale,
    cssClass: 'axis-bottom',
    tickFormat: d => {
      let date = new Date(d * 1000)
      return ('0' + date.getHours()).slice(-2) + ':' +
        ('0' + date.getMinutes()).slice(-2) + ':' + ('0' + date.getSeconds()).slice(-2)
    }
  }
  const yAxisProps = {
    orient: 'Left',
    translate: `translate(${margins.left}, 0)`,
    scale: scales.yScale,
    cssClass: 'axis-left',
    tickFormat: null
  }

  return (
    <g>
      <Axis {...xAxisProps} />
      <Axis {...yAxisProps} />
    </g>
  )
}

const Area = ({ scales, margins, dimensions, data }) => {
  const { xScale, yScale } = scales;
  const _area = d3.area()
    .x(data => xScale(data.timestamp))
    .y0(data => yScale(data.val))
    .y1(dimensions.height - margins.bottom)
    .curve(d3.curveMonotoneX)
  const _line = key => d3.line()
    .x(data => xScale(data.timestamp))
    .y(data => yScale(data[key]))
    .curve(d3.curveMonotoneX)
  const area = <path className="area" d={_area(data)} />
  const thresholdLine = <path className="threshold-line" d={_line('threshold')(data)} />
  const dataLine = <path className="data-line" d={_line('val')(data)} />

  return (
    <g>
      {dataLine}{thresholdLine}{area}
    </g>
  )
}

const Tooltip = ({ scales, margins, dimensions, data, flameGraphHandler, hasFlameGraphData }) => {
  const { xScale, yScale } = scales
  const getDataIdx = d3.bisector(d => d.timestamp).left
  const getData = offsetX => {
    const mouseVal = xScale.invert(offsetX);
    const idx = getDataIdx(data, mouseVal)

    return (mouseVal - data[idx - 1].timestamp) < (data[idx].timestamp - mouseVal) ? data[idx - 1] : data[idx]
  }
  const tooltip = (
    <g className="tooltip" transform={`translate(${xScale(data[0].timestamp)}, ${yScale(data[0].val)})`}>
      <line y1="0" y2={dimensions.height - margins.bottom} stroke="steelblue"
        strokeWidth="1px" strokeDasharray="5" />
      <circle r="6px" stroke="steelblue" strokeWidth="3px" fill="#333333" />
      <text x="-10" y="-10" fontSize="12px">
        {data[0].val}
      </text>
    </g>
  )
  const overlay = (
    <rect
      transform={`translate(${margins.left}, ${margins.top})`}
      width={dimensions.width - margins.left - margins.right}
      height={dimensions.height - margins.top - margins.bottom}
      opacity="0"
      onMouseMove={event => {
        if (hasFlameGraphData()) {
          return
        }
        const d = getData(event.nativeEvent.offsetX)
        d3.select(".tooltip").attr("transform", "translate(" + xScale(d.timestamp) + ", " + yScale(d.val) + ")");
        d3.select(".tooltip line").attr("y2", dimensions.height - yScale(d.val) - margins.bottom);
        d3.select(".tooltip text").text(d.val)
      }}
      onMouseDown={event => {
        if (hasFlameGraphData()) {
          return
        }
        const d = getData(event.nativeEvent.offsetX)
        flameGraphHandler(d)
      }}
    />
  )

  console.log(data)
  return (
    <g>
      {overlay}{tooltip}
    </g>
  );
}

const MemoryViewChart = ({ margins, dimensions, data, flameGraphHandler, hasFlameGraphData }) => {
  const xScale = d3.scaleLinear()
    .domain(d3.extent(data, d => d.timestamp))
    .range([margins.left, dimensions.width - margins.right])
  const yScale = d3.scaleLinear()
    .domain([0, d3.max(data, d => Math.max(d.threshold, d.val))])
    .range([dimensions.height - margins.top, margins.bottom])
  const text = (
    <text transform="translate(40,140)rotate(-90)" fontSize="13">
      Memory Free(kB)
    </text>
  )
  const rectOverlay = (
    <rect transform={`translate(${margins.left / 2}, ${margins.top / 2})`}
      className="rect-overlay"
      width={dimensions.width - margins.right}
      height={dimensions.height - margins.top} rx="5" ry="5" />
  )
  return (
    <svg width={dimensions.width} height={dimensions.height}>
      {rectOverlay}{text}
      <XYAxisChart scales={{ xScale, yScale }} margins={margins} dimensions={dimensions} data={data} />
      <Area scales={{ xScale, yScale }} margins={margins} dimensions={dimensions} data={data} />
      <Tooltip scales={{ xScale, yScale }} margins={margins} dimensions={dimensions} data={data}
        flameGraphHandler={flameGraphHandler} hasFlameGraphData={hasFlameGraphData} />
    </svg>
  )
}

const MemleakProfileView = () => {
  const [data, setData] = useState()
  const [flameGraphData, setFlameGraphData] = useState()
  const hasFlameGraphData = () => {
    return !!flameGraphData
  }

  useEffect(() => {
    d3.json('/' + GROUP + '/' + ROUTINE).then(data => {
      setData(data)
    })
  }, [])

  const margins = { top: 50, right: 100, bottom: 50, left: 100 },
    dimensions = { height: screen.height / 2, width: screen.width / 2 };

  if (!data) {
    return (
      <div>
        Loading...
      </div>
    )
  }
  return (
    <div>
      <MemoryViewChart className="overview-chart" margins={margins} dimensions={dimensions} data={data}
        flameGraphHandler={setFlameGraphData} hasFlameGraphData={hasFlameGraphData} />
      {flameGraphData && <FlameGraph timestamp={flameGraphData.timestamp} group={GROUP}
        routine={ROUTINE} closeHandler={() => { setFlameGraphData(null) }} />}
    </div>
  )
}

export default MemleakProfileView
