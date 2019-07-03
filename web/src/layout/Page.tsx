import React, {ReactNode} from "react";

export class Page extends React.Component<any> {
	componentWillMount(): void {
		document.title = this.props.title
	}

	render(): ReactNode {
		return (
			<main id="main">
				<div id="content-wrap">
					<div id="content">{this.props.children}</div>
				</div>
			</main>
		);
	}
}