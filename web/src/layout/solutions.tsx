import React, {FC, FormEventHandler} from "react";
import {Block, BlockProps, FormBlock} from "./blocks";
import {Button} from "./buttons";
import {Compiler, getDefense, getShortVerdict, Solution} from "../api";
import Input from "./Input";
import {Link} from "react-router-dom";
import "./solutions.scss";

export type SubmitSolutionSideBlockProps = {
	onSubmit: FormEventHandler;
	compilers?: Compiler[];
};

export const SubmitSolutionSideBlock: FC<SubmitSolutionSideBlockProps> = props => {
	const {onSubmit, compilers} = props;
	return <FormBlock onSubmit={onSubmit} title="Submit solution" footer={
		<Button color="primary">Submit</Button>
	}>
		<div className="ui-field">
			<label>
				<span className="label">Compiler:</span>
				<select name="compilerID">
					{compilers && compilers.map((compiler, index) =>
						<option value={compiler.ID} key={index}>{compiler.Name}</option>
					)}
				</select>
			</label>
			<label>
				<span className="label">Source file:</span>
				<Input type="file" name="sourceFile" placeholder="Source code"/>
			</label>
		</div>
	</FormBlock>;
};

export type SolutionsSideBlockProps = {
	solutions: Solution[];
};

export const SolutionsSideBlock: FC<SolutionsSideBlockProps> = props => {
	const {solutions} = props;
	return <Block title="Solutions">
		<ul>{solutions && solutions.map(
			(solution, index) => <li key={index}>
				<Link to={"/solutions/" + solution.ID}>{solution.ID}</Link>
				{solution.Report && <span className="verdict">{getShortVerdict(solution.Report.Verdict)}</span>}
			</li>
		)}</ul>
	</Block>
};

export type SolutionsBlockProps = BlockProps & {
	solutions: Solution[];
};

export const SolutionsBlock: FC<SolutionsBlockProps> = props => {
	let {solutions, className, ...rest} = props;
	className = className ? "b-solutions " + className : "b-solutions";
	const format = (n: number) => {
		return ("0" + n).slice(-Math.max(2, String(n).length));
	};
	const formatDate = (d: Date) =>
		[d.getFullYear(), d.getMonth() + 1, d.getDate()].map(format).join("-");
	const formatTime = (d: Date) =>
		[d.getHours(), d.getMinutes(), d.getSeconds()].map(format).join(":");
	return <Block className={className} {...rest}>
		<table className="ui-table">
			<thead>
			<tr>
				<th className="id">#</th>
				<th className="created">Created</th>
				<th className="participant">Participant</th>
				<th className="problem">Problem</th>
				<th className="verdict">Verdict</th>
				<th className="defense">Defense</th>
			</tr>
			</thead>
			<tbody>
			{solutions && solutions.map((solution, index) => {
				const {ID, CreateTime, User, Problem, Report} = solution;
				let createDate = new Date(CreateTime * 1000);
				return <tr key={index} className="solution">
					<td className="id">
						<Link to={"/solutions/" + ID}>{ID}</Link>
					</td>
					<td className="created">
						<div className="time">{formatTime(createDate)}</div>
						<div className="date">{formatDate(createDate)}</div>
					</td>
					<td className="participant">{User ?
						<Link to={"/users/" + User.Login}>{User.Login}</Link> :
						<>&mdash;</>
					}</td>
					<td className="problem">{Problem ?
						<Link to={"/problems/" + Problem.ID}>{Problem.Title}</Link> :
						<span>&mdash;</span>
					}</td>
					<td className="verdict">
						<div className="type">{Report && getShortVerdict(Report.Verdict)}</div>
						<div className="value">{Report && Report.Data.Points}</div>
					</td>
					<td className="defense">
						{Report && getDefense(Report.Data.Defense)}
					</td>
				</tr>;
			})}
			</tbody>
		</table>
	</Block>;
};
