/// <reference path="../References.d.ts"/>
import Dispatcher from '../dispatcher/Dispatcher';
import EventEmitter from '../EventEmitter';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as GlobalTypes from '../types/GlobalTypes';

class OrganizationsStore extends EventEmitter {
	_organizations: OrganizationTypes.OrganizationsRo = Object.freeze([]);
	_map: {[key: string]: number} = {};
	_token = Dispatcher.register((this._callback).bind(this));

	get organizations(): OrganizationTypes.OrganizationsRo {
		return this._organizations;
	}

	get organizationsM(): OrganizationTypes.Organizations {
		let organizations: OrganizationTypes.Organizations = [];
		this._organizations.forEach((
				organization: OrganizationTypes.OrganizationRo): void => {
			organizations.push({
				...organization,
			});
		});
		return organizations;
	}

	organization(id: string): OrganizationTypes.OrganizationRo {
		let i = this._map[id];
		if (i === undefined) {
			return null;
		}
		return this._organizations[i];
	}

	emitChange(): void {
		this.emitDefer(GlobalTypes.CHANGE);
	}

	addChangeListener(callback: () => void): void {
		this.on(GlobalTypes.CHANGE, callback);
	}

	removeChangeListener(callback: () => void): void {
		this.removeListener(GlobalTypes.CHANGE, callback);
	}

	_sync(organizations: OrganizationTypes.Organization[]): void {
		this._map = {};
		for (let i = 0; i < organizations.length; i++) {
			organizations[i] = Object.freeze(organizations[i]);
			this._map[organizations[i].id] = i;
		}

		this._organizations = Object.freeze(organizations);
		this.emitChange();
	}

	_callback(action: OrganizationTypes.OrganizationDispatch): void {
		switch (action.type) {
			case OrganizationTypes.SYNC:
				this._sync(action.data.organizations);
				break;
		}
	}
}

export default new OrganizationsStore();