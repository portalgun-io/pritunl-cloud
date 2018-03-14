/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';
import * as VpcActions from '../actions/VpcActions';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as OrganizationTypes from "../types/OrganizationTypes";
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	vpc: VpcTypes.VpcRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addNetworkRole: string;
	addVpc: string;
	vpc: VpcTypes.Vpc;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
		backgroundColor: 'rgba(0, 0, 0, 0.13)',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '250px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class VpcDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addNetworkRole: null,
			addVpc: null,
			vpc: null,
		};
	}

	set(name: string, val: any): void {
		let vpc: any;

		if (this.state.changed) {
			vpc = {
				...this.state.vpc,
			};
		} else {
			vpc = {
				...this.props.vpc,
			};
		}

		vpc[name] = val;

		this.setState({
			...this.state,
			changed: true,
			vpc: vpc,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		VpcActions.commit(this.state.vpc).then((): void => {
			this.setState({
				...this.state,
				message: 'Your changes have been saved',
				changed: false,
				disabled: false,
			});

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						vpc: null,
						changed: false,
					});
				}
			}, 1000);

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						message: '',
					});
				}
			}, 3000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		VpcActions.remove(this.props.vpc.id).then((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let vpc: VpcTypes.Vpc = this.state.vpc ||
			this.props.vpc;

		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
			datacentersSelect.push(
				<option key="null" value="">Node Vpc</option>);

			for (let datacenter of this.props.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
		}

		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			organizationsSelect.push(
				<option key="null" value="">Node Vpc</option>);

			for (let organization of this.props.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

		return <td
			className="pt-cell"
			colSpan={5}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('open-ignore') !== -1) {
								return;
							}

							this.props.onClose();
						}}
					>
            <div>
              <label
                className="pt-control pt-checkbox open-ignore"
                style={css.select}
              >
                <input
                  type="checkbox"
                  className="open-ignore"
                  checked={this.props.selected}
                  onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
                />
                <span className="pt-control-indicator open-ignore"/>
              </label>
            </div>
						<div className="flex"/>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-trash open-ignore"
							style={css.button}
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm vpc remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of vpc"
						type="text"
						placeholder="Enter name"
						value={vpc.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageInput
						label="Network"
						help="Network address of vpc with cidr."
						type="text"
						placeholder="Enter network"
						value={vpc.network}
						onChange={(val): void => {
							this.set('network', val);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.vpc.id || 'Unknown',
							},
						]}
					/>
					<PageSelect
						disabled={this.state.disabled}
						label="Organization"
						help="Organization for vpc."
						value={vpc.organization}
						onChange={(val): void => {
							this.set('organization', val);
						}}
					>
						{organizationsSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Datacenter"
						help="Datacenter for vpc."
						value={vpc.datacenter}
						onChange={(val): void => {
							this.set('datacenter', val);
						}}
					>
						{datacentersSelect}
					</PageSelect>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.vpc && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						vpc: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}