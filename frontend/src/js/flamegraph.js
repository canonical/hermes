import React, {useEffect} from 'react'
import * as d3 from 'd3'
import {flamegraph} from 'd3-flame-graph'
import "../css/flamegraph.scss"
import '../../node_modules/d3-flame-graph/dist/d3-flamegraph.css'

const FlameGraph = ({timestamp, category, stackFile, closeHandler}) => {
	const chart = <div id='chart'></div>
	const date = new Date(timestamp * 1000)
	const title = ('0' + date.getMonth()).slice(-2) + '/' +
		('0' + date.getDay()).slice(-2) + ' ' + ('0' + date.getHours()).slice(-2) + ':' +
		('0' + date.getMinutes()).slice(-2) + ':' + ('0' + date.getSeconds()).slice(-2)
	const flameGraph = flamegraph()
		.width(1460)
		.cellHeight(18)
		.transitionDuration(750)
		.transitionEase(d3.easeCubic)
		.sort(true)
		.selfValue(false)

	useEffect(() => {
		d3.json("/view/" + category + "/" + timestamp.toString() + "/" + stackFile).then(data => {
				d3.select("#chart")
					.datum(data)
					.call(flameGraph);
		})
	}, [])

	return (
		<div className='box'>
			<div className='title'>
				{title}
			</div>
			<span className='close-icon' onClick={closeHandler}>x</span>
			<button className='reset_zoom' onClick={() => flameGraph.resetZoom()}>Reset zoom</button>
			{chart}
		</div>
	)
}

export default FlameGraph
