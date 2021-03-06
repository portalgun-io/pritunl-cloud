package block

import (
	"fmt"
	"net"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Block struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Subnets  []string           `bson:"subnets" json:"subnets"`
	Excludes []string           `bson:"excludes" json:"excludes"`
	Netmask  string             `bson:"netmask" json:"netmask"`
	Gateway  string             `bson:"gateway" json:"gateway"`
}

func (b *Block) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if b.Subnets == nil {
		b.Subnets = []string{}
	}

	if b.Excludes == nil {
		b.Excludes = []string{}
	}

	if b.Gateway != "" {
		gateway := net.ParseIP(b.Gateway)
		if gateway == nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_gateway",
				Message: "Gateway address is invalid",
			}
			return
		}
	}

	if b.Netmask != "" {
		netmask := utils.ParseIpMask(b.Netmask)
		if netmask == nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_netmask",
				Message: "Netmask is invalid",
			}
			return
		}
	}

	subnets := []string{}
	for _, subnet := range b.Subnets {
		if !strings.Contains(subnet, "/") {
			subnet += "/32"
		}

		_, subnetNet, e := net.ParseCIDR(subnet)
		if e != nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_subnet",
				Message: "Invalid subnet address",
			}
			return
		}

		subnets = append(subnets, subnetNet.String())
	}
	b.Subnets = subnets

	excludes := []string{}
	for _, exclude := range b.Excludes {
		if !strings.Contains(exclude, "/") {
			exclude += "/32"
		}

		_, excludeNet, e := net.ParseCIDR(exclude)
		if e != nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_exclude",
				Message: "Invalid exclude address",
			}
			return
		}

		excludes = append(excludes, excludeNet.String())
	}
	b.Excludes = excludes

	return
}

func (b *Block) GetGateway() net.IP {
	return net.ParseIP(b.Gateway)
}

func (b *Block) GetMask() net.IPMask {
	return utils.ParseIpMask(b.Netmask)
}

func (b *Block) GetGatewayCidr() string {
	staticGateway := net.ParseIP(b.Gateway)
	staticMask := utils.ParseIpMask(b.Netmask)
	if staticGateway == nil || staticMask == nil {
		return ""
	}

	staticSize, _ := staticMask.Size()
	return fmt.Sprintf("%s/%d", staticGateway.String(), staticSize)
}

func (b *Block) GetNetwork() (staticNet *net.IPNet, err error) {
	staticMask := utils.ParseIpMask(b.Netmask)
	if staticMask == nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Invalid netmask"),
		}
		return
	}
	staticSize, _ := staticMask.Size()

	_, staticNet, err = net.ParseCIDR(
		fmt.Sprintf("%s/%d", b.Gateway, staticSize))
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Failed to parse network cidr"),
		}
		return
	}

	return
}

func (b *Block) GetIps(db *database.Database) (blckIps set.Set, err error) {
	coll := db.BlocksIp()

	ipsInf, err := coll.Distinct(db, "ip", &bson.M{
		"block": b.Id,
	})

	blckIps = set.NewSet()
	for _, ipInf := range ipsInf {
		if ip, ok := ipInf.(int64); ok {
			blckIps.Add(ip)
		}
	}

	return
}

func (b *Block) GetIp(db *database.Database,
	instId primitive.ObjectID, typ string) (ip net.IP, err error) {

	blckIps, err := b.GetIps(db)
	if err != nil {
		return
	}

	coll := db.BlocksIp()
	gateway := net.ParseIP(b.Gateway)
	if gateway == nil {
		err = &errortypes.ParseError{
			errors.New("block: Failed to parse block gateway"),
		}
		return
	}

	gatewaySize, _ := b.GetMask().Size()
	_, gatewayCidr, err := net.ParseCIDR(fmt.Sprintf("%s/%d",
		gateway.String(), gatewaySize))
	if err != nil {
		err = &errortypes.ParseError{
			errors.New("block: Failed to parse block gateway cidr"),
		}
		return
	}

	broadcast := utils.GetLastIpAddress(gatewayCidr)

	excludes := []*net.IPNet{}
	for _, exclude := range b.Excludes {
		_, network, e := net.ParseCIDR(exclude)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block exclude"),
			}
			return
		}

		excludes = append(excludes, network)
	}

	for _, subnet := range b.Subnets {
		_, network, e := net.ParseCIDR(subnet)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block subnet"),
			}
			return
		}

		first := true
		curIp := utils.CopyIpAddress(network.IP)
		for {
			if first {
				first = false
			} else {
				utils.IncIpAddress(curIp)
			}
			curIpInt := utils.IpAddress2Int(curIp)

			if !network.Contains(curIp) {
				break
			}

			if blckIps.Contains(curIpInt) || gatewayCidr.IP.Equal(curIp) ||
				gateway.Equal(curIp) || broadcast.Equal(curIp) {

				continue
			}

			excluded := false
			for _, exclude := range excludes {
				if exclude.Contains(curIp) {
					excluded = true
					break
				}
			}

			if excluded {
				continue
			}

			blckIp := &BlockIp{
				Block:    b.Id,
				Ip:       utils.IpAddress2Int(curIp),
				Instance: instId,
				Type:     typ,
			}

			_, err = coll.InsertOne(db, blckIp)
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.DuplicateKeyError); ok {
					err = nil
					continue
				}
				return
			}

			ip = curIp
			break
		}

		if ip != nil {
			break
		}
	}

	if ip == nil {
		err = &BlockFull{
			errors.New("block: Address pool full"),
		}
		return
	}

	return
}

func (b *Block) RemoveIp(db *database.Database,
	instId primitive.ObjectID) (err error) {

	coll := db.BlocksIp()
	_, err = coll.DeleteMany(db, &bson.M{
		"instance": instId,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}

func (b *Block) ValidateAddresses(db *database.Database,
	commitFields set.Set) (err error) {

	coll := db.Blocks()
	ipColl := db.BlocksIp()
	instColl := db.Instances()

	gateway := net.ParseIP(b.Gateway)
	excludes := []*net.IPNet{}
	for _, exclude := range b.Excludes {
		_, network, e := net.ParseCIDR(exclude)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block exclude"),
			}
			return
		}

		excludes = append(excludes, network)
	}

	subnets := []*net.IPNet{}
	for _, subnet := range b.Subnets {
		_, network, e := net.ParseCIDR(subnet)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block subnet"),
			}
			return
		}

		subnets = append(subnets, network)
	}

	if commitFields != nil {
		err = coll.CommitFields(b.Id, b, commitFields)
		if err != nil {
			return
		}
	}

	cursor, err := ipColl.Find(db, &bson.M{
		"block": b.Id,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		blckIp := &BlockIp{}
		err = cursor.Decode(blckIp)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		remove := false
		ip := utils.Int2IpAddress(blckIp.Ip)

		if gateway != nil && gateway.Equal(ip) {
			remove = true
		}

		if !remove {
			for _, exclude := range excludes {
				if exclude.Contains(ip) {
					remove = true
					break
				}
			}
		}

		if !remove {
			match := false
			for _, subnet := range subnets {
				if subnet.Contains(ip) {
					match = true
					break
				}
			}

			if !match {
				remove = true
			}
		}

		if remove {
			_, _ = instColl.UpdateOne(db, &bson.M{
				"_id": blckIp.Instance,
			}, &bson.M{
				"$set": &bson.M{
					"restart_block_ip": true,
				},
			})

			_, err = ipColl.DeleteOne(db, &bson.M{
				"_id": blckIp.Id,
			})
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (b *Block) Commit(db *database.Database) (err error) {
	coll := db.Blocks()

	err = coll.Commit(b.Id, b)
	if err != nil {
		return
	}

	return
}

func (b *Block) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	err = b.ValidateAddresses(db, fields)
	if err != nil {
		return
	}

	return
}

func (b *Block) Insert(db *database.Database) (err error) {
	coll := db.Blocks()

	if !b.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("block: Block already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, b)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func init() {
	module := requires.New("block")
	module.After("settings")

	module.Handler = func() (err error) {
		db := database.GetDatabase()
		defer db.Close()

		coll := db.BlocksIp()

		// TODO Upgrade <= 1.0.1173.24
		_, err = coll.UpdateMany(db, &bson.M{
			"type": &bson.M{
				"$exists": false,
			},
		}, &bson.M{
			"$set": &bson.M{
				"type": External,
			},
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}

		return
	}
}
