package spatialmath

import (
	"math"
	"testing"

	"github.com/golang/geo/r3"
	"go.viam.com/test"

	"go.viam.com/rdk/utils"
)

func makeTestMesh(o Orientation, pt r3.Vector, triangles []*Triangle) *Mesh {
	return NewMesh(NewPose(pt, o), triangles, "")
}

func makeSimpleTriangleMesh() *Mesh {
	// Create a simple triangle mesh at origin
	tri1 := NewTriangle(
		r3.Vector{X: 0, Y: 0, Z: 0},
		r3.Vector{X: 1, Y: 0, Z: 0},
		r3.Vector{X: 0, Y: 1, Z: 0},
	)
	tri2 := NewTriangle(
		r3.Vector{X: 0.6, Y: 0.6, Z: 0},
		r3.Vector{X: 1, Y: 0, Z: 0},
		r3.Vector{X: 0, Y: 1, Z: 0},
	)
	tri3 := NewTriangle(
		r3.Vector{X: 0, Y: 0, Z: 10},
		r3.Vector{X: 1, Y: 0, Z: 10},
		r3.Vector{X: 0, Y: 1, Z: 10},
	)
	return makeTestMesh(NewZeroOrientation(), r3.Vector{}, []*Triangle{tri1, tri2, tri3})
}

func TestNewMesh(t *testing.T) {
	tri := NewTriangle(
		r3.Vector{X: 0, Y: 0, Z: 0},
		r3.Vector{X: 1, Y: 0, Z: 0},
		r3.Vector{X: 0, Y: 1, Z: 0},
	)
	pose := NewPose(r3.Vector{X: 1, Y: 2, Z: 3}, NewZeroOrientation())

	mesh := NewMesh(pose, []*Triangle{tri}, "test_mesh")

	test.That(t, mesh.Label(), test.ShouldEqual, "test_mesh")
	test.That(t, PoseAlmostEqual(mesh.Pose(), pose), test.ShouldBeTrue)
	test.That(t, len(mesh.Triangles()), test.ShouldEqual, 1)
}

func TestMeshProtoConversion(t *testing.T) {
	m, err := NewMeshFromPLYFile(utils.ResolveFile("spatialmath/data/simple.ply"))
	test.That(t, err, test.ShouldBeNil)
	m2, err := NewGeometryFromProto(m.ToProtobuf())
	test.That(t, err, test.ShouldBeNil)

	test.That(t, PoseAlmostEqual(m.Pose(), m2.Pose()), test.ShouldBeTrue)
	test.That(t, m.Label(), test.ShouldResemble, m2.Label())
	test.That(t, len(m.Triangles()), test.ShouldEqual, 2)
	test.That(t, len(m2.(*Mesh).Triangles()), test.ShouldEqual, 2)
	test.That(t, m.Triangles()[0], test.ShouldResemble, m2.(*Mesh).Triangles()[0])
	test.That(t, m.Triangles()[1], test.ShouldResemble, m2.(*Mesh).Triangles()[1])
}

func TestMeshTransform(t *testing.T) {
	mesh := makeSimpleTriangleMesh()

	// Transform mesh by translation
	newPose := NewPose(r3.Vector{X: 1, Y: 0, Z: 0}, NewZeroOrientation())
	transformed := mesh.Transform(newPose)

	// Check that transformed mesh has correct pose
	test.That(t, transformed.Pose().Point().X, test.ShouldEqual, 1)

	// Original mesh should be unchanged
	test.That(t, mesh.Pose().Point().X, test.ShouldEqual, 0)
}

func TestMeshCollidesWithMesh(t *testing.T) {
	mesh1 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
		[]*Triangle{NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1, Z: 0},
		)})

	// A mesh has 3 parts: {vertex, edge, face} ==> 6 possible basic collisions, accounting for symmetry

	// vertex-vertex
	t.Run("triangle vertex against triangle vertex", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0, Y: 0, Z: 0},
				r3.Vector{X: 0, Y: 0, Z: 1},
				r3.Vector{X: 1, Y: 0, Z: 1},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// vertex-edge
	t.Run("triangle vertex against triangle edge", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0, Y: 0, Z: 1},
				r3.Vector{X: 0, Y: 0, Z: -1},
				r3.Vector{X: -1, Y: 0, Z: 0},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// vertex-face
	t.Run("triangle vertex against triangle face", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0.4, Y: 0.4, Z: 0},
				r3.Vector{X: 0, Y: 0.4, Z: 1},
				r3.Vector{X: 1, Y: 0.4, Z: 1},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// edge-edge
	t.Run("triangle edge against triangle edge", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0.5, Y: 0, Z: 0.5},
				r3.Vector{X: 0.5, Y: 0, Z: -0.5},
				r3.Vector{X: 0.5, Y: -1, Z: 0},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// edge-face. This implies one of the above collision types (e.g., e-e)
	// so if they're all perfectly tested (difficult to guarantee) we're fine
	// nonetheless worth keeping: e-f is the basic collision type checked by collidesWithMesh,
	// and the special case of e parallel to f is important
	t.Run("triangle edge against triangle face", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0.5, Y: -0.1, Z: 0},
				r3.Vector{X: -0.1, Y: 0.5, Z: 0},
				r3.Vector{X: 0, Y: 0, Z: 1},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// face-face. This implies one of the above collision types
	t.Run("triangle face against triangle face", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0.5, Y: -0.1, Z: 0},
				r3.Vector{X: -0.1, Y: 0.5, Z: 0},
				r3.Vector{X: 0.6, Y: 0.6, Z: 0},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Test collision with no edge intersections
	t.Run("clipped triangles", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0.5, Y: 0.1, Z: 0.5},
				r3.Vector{X: 0.5, Y: 0.1, Z: -0.5},
				r3.Vector{X: -1, Y: 0, Z: 0},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Test collision with non-overlapping mesh
	t.Run("non-overlapping triangles", func(t *testing.T) {
		mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0, Y: 0, Z: 0.2},
				r3.Vector{X: 1, Y: 0, Z: 0.5},
				r3.Vector{X: 0, Y: 1, Z: 0.3},
			)})
		collides, err := mesh1.CollidesWith(mesh2, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeFalse)
	})
}

func TestMeshCollidesWithCapsule(t *testing.T) {
	mesh := makeTestMesh(NewZeroOrientation(), r3.Vector{},
		[]*Triangle{NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1, Z: 0},
		)})

	// A mesh has 3 parts: {vertex, edge, face}
	// A capsule has approx 3 parts: {extreme point, general spherical point, cylinder point}
	// We enumerate the 9 possible pairs

	// Collision with triangle vertex
	// Capsule extreme vertex collision (with triangle vertex)
	t.Run("triangle vertex against capsule endpoint", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: 0, Y: 0, Z: 1.5},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Capsule non-extreme spherical point collision (with triangle vertex)
	t.Run("triangle vertex against capsule generic spherical point", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: -0.75, Y: 0, Z: 1},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Capsule cylinder point collision (with triangle vertex)
	t.Run("triangle vertex against capsule cylinder point", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: -1, Y: 0, Z: 0},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Collision with triangle edge
	// Capsule extreme vertex collision (with triangle edge)
	t.Run("triangle edge against capsule endpoint", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: 0.5, Y: 0, Z: 1.5},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Capsule non-extreme spherical vertex collision (with triangle edge)
	t.Run("triangle edge against capsule generic spherical point", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: -0.75, Y: 0.5, Z: 1},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Capsule cylinder vertex collision (with triangle edge)
	t.Run("triangle edge against capsule cylinder point", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: 0.5, Y: -1, Z: 0},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Collision with triangle face
	// Capsule extreme vertex collision (with triangle face)
	t.Run("triangle face against capsule endpoint", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: 0.5, Y: 0.5, Z: 1.5},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Capsule non-extreme spherical vertex collision (with triangle face)
	t.Run("triangle face against capsule generic spherical point", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: 0.5, Y: 0.5, Z: 1 + math.Sqrt(2)/4},
			&OrientationVector{OY: 1, OZ: 1}), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Capsule cylinder vertex collision (with triangle face) point collision not possible, have to use a line
	t.Run("triangle face against capsule cylinder point", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: 0.2, Y: 0.2, Z: 0.1},
			&OrientationVector{OX: 1}), 0.1, 0.3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Partially encompassing capsule (could potentially divide into more cases, but this (only face collisions) should be most restrictive)
	t.Run("capsule encompassing triangle face", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: 0.2, Y: 0.2, Z: 0},
			NewZeroOrientation()), 0.1, 0.3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Completely encompassing capsule, no boundary collision
	t.Run("capsule completely encompassing triangle", func(t *testing.T) {
		capsule, err := NewCapsule(NewZeroPose(), 2, 4.5, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Non-overlapping capsule
	t.Run("capsule not touching triangle", func(t *testing.T) {
		capsule, err := NewCapsule(NewPose(r3.Vector{X: -1.1, Y: -1.1, Z: 0},
			NewZeroOrientation()), 1, 3, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(capsule, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeFalse)
	})
}

func TestMeshCollidesWithBox(t *testing.T) {
	mesh := makeSimpleTriangleMesh()
	// Types of triangle points: {vertex, edge, face}
	// Types of box points: {vertex, edge, face}
	// We exhaust the 9 collision options

	// Collision with triangle vertex
	// Box vertex collision (with triangle vertex)
	t.Run("Box vertex against triangle vertex", func(t *testing.T) {
		box, err := NewBox(NewPose(r3.Vector{X: 1.5, Y: 0.5, Z: 0.5}, NewZeroOrientation()),
			r3.Vector{X: 1, Y: 1, Z: 1}, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Box edge collision (with triangle vertex)
	t.Run("Box edge against triangle vertex", func(t *testing.T) {
		box, err := NewBox(NewPose(r3.Vector{X: 1.5, Y: 0, Z: 0.5}, NewZeroOrientation()),
			r3.Vector{X: 1, Y: 1, Z: 1}, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Box face collision (with triangle vertex)
	t.Run("Box face against triangle vertex", func(t *testing.T) {
		box, err := NewBox(NewPose(r3.Vector{X: 1.5, Y: 0, Z: 0}, NewZeroOrientation()),
			r3.Vector{X: 1, Y: 1, Z: 1}, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Collision with triangle edge
	// Box vertex collision (with triangle edge)
	t.Run("Box vertex against triangle edge", func(t *testing.T) {
		box, err := NewBox(NewPose(r3.Vector{X: 0.7, Y: 1.5 - 0.7*(3.0/2), Z: 0.5}, NewZeroOrientation()), // idk how to do orientation loool
			r3.Vector{X: 1, Y: 1, Z: 1}, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Box edge collision (with triangle edge)
	t.Run("Box edge against triangle edge", func(t *testing.T) {
		box, err := NewBox(NewPose(r3.Vector{X: 0.5, Y: -math.Sqrt(2) / 2, Z: 0},
			&OrientationVector{Theta: math.Pi / 4}), r3.Vector{X: 1, Y: 1, Z: 1}, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Partially encompassing box, no triangle vertices inside the box
	t.Run("Box clipping triangle", func(t *testing.T) {
		box, err := NewBox(NewPose(r3.Vector{X: 0.9, Y: 0.9, Z: 0}, NewZeroOrientation()),
			r3.Vector{X: 1, Y: 1, Z: 1}, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Completely encompassing box, no boundary collision
	t.Run("Box strictly encompassing triangle", func(t *testing.T) {
		box, err := NewBox(NewZeroPose(),
			r3.Vector{X: 4, Y: 4, Z: 4}, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Create non-overlapping box
	t.Run("Box not touching triangle", func(t *testing.T) {
		box, err := NewBox(NewPose(r3.Vector{X: 2, Y: 2, Z: 2}, NewZeroOrientation()),
			r3.Vector{X: 1, Y: 1, Z: 1}, "")
		test.That(t, err, test.ShouldBeNil)

		collides, err := mesh.CollidesWith(box, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeFalse)
	})
}

func TestMeshCollidesWithPoint(t *testing.T) {
	mesh := makeTestMesh(NewZeroOrientation(), r3.Vector{},
		[]*Triangle{NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1, Z: 0},
		)})

	// Collision with triangle vertex
	t.Run("Point against triangle vertex", func(t *testing.T) {
		point := NewPoint(r3.Vector{}, "")
		collides, err := mesh.CollidesWith(point, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Collision with triangle edge
	t.Run("Point against triangle edge", func(t *testing.T) {
		point := NewPoint(r3.Vector{X: 0, Y: 0.5, Z: 0}, "")
		collides, err := mesh.CollidesWith(point, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Collision with triangle face
	t.Run("Point against triangle face", func(t *testing.T) {
		point := NewPoint(r3.Vector{X: 0.3, Y: 0.3, Z: 0}, "")
		collides, err := mesh.CollidesWith(point, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Point not touching triangle
	t.Run("Point not touching triangle", func(t *testing.T) {
		point := NewPoint(r3.Vector{X: 0, Y: 0, Z: 2 * defaultCollisionBufferMM}, "")
		collides, err := mesh.CollidesWith(point, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeFalse)
	})
}

func TestMeshCollidesWithSphere(t *testing.T) {
	mesh := makeTestMesh(NewZeroOrientation(), r3.Vector{},
		[]*Triangle{NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1, Z: 0},
		)})

	// Collision with triangle vertex
	t.Run("Sphere against triangle vertex", func(t *testing.T) {
		sphere, err := NewSphere(NewPose(r3.Vector{X: 0, Y: 0, Z: 1}, NewZeroOrientation()), 1, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(sphere, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Collision with triangle edge
	t.Run("Sphere against triangle edge", func(t *testing.T) {
		sphere, err := NewSphere(NewPose(r3.Vector{X: 0.5, Y: 0, Z: 1}, NewZeroOrientation()), 1, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(sphere, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Collision with triangle face
	t.Run("Sphere against triangle face", func(t *testing.T) {
		sphere, err := NewSphere(NewPose(r3.Vector{X: 0.3, Y: 0.3, Z: 1}, NewZeroOrientation()), 1, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(sphere, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Sphere clipping triangle
	t.Run("Sphere clipping triangle", func(t *testing.T) {
		sphere, err := NewSphere(NewPose(r3.Vector{X: 0.3, Y: 0.3, Z: 0}, NewZeroOrientation()), 0.1, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(sphere, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Sphere completely encompassing triangle
	t.Run("Sphere completely encompassing triangle", func(t *testing.T) {
		sphere, err := NewSphere(NewZeroPose(), 2, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(sphere, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeTrue)
	})

	// Sphere not touching triangle
	t.Run("Sphere not touching triangle", func(t *testing.T) {
		sphere, err := NewSphere(NewPose(r3.Vector{X: 0, Y: 0, Z: 1 + 2*defaultCollisionBufferMM}, NewZeroOrientation()), 1, "")
		test.That(t, err, test.ShouldBeNil)
		collides, err := mesh.CollidesWith(sphere, defaultCollisionBufferMM)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, collides, test.ShouldBeFalse)
	})
}

func TestMeshDistanceFrom(t *testing.T) {
	mesh1 := makeSimpleTriangleMesh()

	// Test distance from overlapping mesh
	mesh2 := makeTestMesh(NewZeroOrientation(), r3.Vector{},
		[]*Triangle{NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1, Z: 0},
		)})

	dist, err := mesh1.DistanceFrom(mesh2)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, dist, test.ShouldEqual, 0)

	// Test distance from non-overlapping mesh
	mesh3 := makeTestMesh(NewZeroOrientation(), r3.Vector{X: 2, Y: 0, Z: 0},
		[]*Triangle{NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1, Z: 0},
		)})

	dist, err = mesh1.DistanceFrom(mesh3)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, dist, test.ShouldBeGreaterThan, 0)
}

func TestMeshToPoints(t *testing.T) {
	t.Run("Simple triangle with density enforced", func(t *testing.T) {
		mesh := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 0, Y: 0, Z: 0},
				r3.Vector{X: 3, Y: 0, Z: 0},
				r3.Vector{X: -3, Y: 3, Z: 0},
			)})

		points := mesh.ToPoints(0.3)

		// Verify points match those expected for similar tiling method
		expectedPoints := []r3.Vector{
			{X: 0, Y: 0, Z: 0},
			{X: 1, Y: 0, Z: 0},
			{X: 2, Y: 0, Z: 0},
			{X: 3, Y: 0, Z: 0},

			{X: -1, Y: 1, Z: 0},
			{X: 0, Y: 1, Z: 0},
			{X: 1, Y: 1, Z: 0},

			{X: -2, Y: 2, Z: 0},
			{X: -1, Y: 2, Z: 0},

			{X: -3, Y: 3, Z: 0},
		}

		test.That(t, len(points), test.ShouldEqual, len(expectedPoints))
		for _, expected := range expectedPoints {
			found := false
			for _, actual := range points {
				if R3VectorAlmostEqual(actual, expected, 1e-10) {
					found = true
					break
				}
			}
			test.That(t, found, test.ShouldBeTrue)
		}
	})

	t.Run("Degenerate triangle", func(t *testing.T) {
		mesh := makeTestMesh(NewZeroOrientation(), r3.Vector{},
			[]*Triangle{NewTriangle(
				r3.Vector{X: 1, Y: 1, Z: 1},
				r3.Vector{X: 1, Y: 1, Z: 1},
				r3.Vector{X: 1, Y: 1, Z: 1},
			)})

		points := mesh.ToPoints(5)

		expectedPoint := r3.Vector{X: 1, Y: 1, Z: 1}

		test.That(t, len(points), test.ShouldEqual, 1)
		test.That(t, points[0], test.ShouldResemble, expectedPoint)
	})
}

func TestMeshEncompassedBy(t *testing.T) {
	mesh := makeSimpleTriangleMesh()

	// Test with encompassing box
	box, err := NewBox(NewZeroPose(), r3.Vector{X: 20, Y: 20, Z: 20}, "")
	test.That(t, err, test.ShouldBeNil)

	encompassed, err := mesh.EncompassedBy(box)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, encompassed, test.ShouldBeTrue)

	// Test with box encompassing some but not all triangles
	smallBox, err := NewBox(NewZeroPose(), r3.Vector{X: 2, Y: 2, Z: 2}, "")
	test.That(t, err, test.ShouldBeNil)

	encompassed, err = mesh.EncompassedBy(smallBox)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, encompassed, test.ShouldBeFalse)
}

func TestMeshProtoConversionFromTriangles(t *testing.T) {
	// Manually create a mesh with a variety of shapes and shared vertices
	triangles := []*Triangle{
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1000, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1000, Z: 0},
		),
		NewTriangle(
			r3.Vector{X: -500, Y: -500, Z: 0},
			r3.Vector{X: 500, Y: -500, Z: 0},
			r3.Vector{X: 0, Y: 500, Z: 0},
		),
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 1000},
			r3.Vector{X: 1000, Y: 0, Z: 1000},
			r3.Vector{X: 500, Y: 1000, Z: 1000},
		),
		NewTriangle(
			r3.Vector{X: 123.456, Y: 789.012, Z: 345.678},
			r3.Vector{X: 456.789, Y: 123.456, Z: 678.901},
			r3.Vector{X: 789.012, Y: 456.789, Z: 123.456},
		),
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 10000, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 10000, Z: 0},
		),
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1, Z: 0},
		),
		NewTriangle(
			r3.Vector{X: -1000, Y: -1000, Z: -1000},
			r3.Vector{X: -500, Y: -1000, Z: -1000},
			r3.Vector{X: -1000, Y: -500, Z: -1000},
		),
		NewTriangle(
			r3.Vector{X: 100, Y: 100, Z: -500},
			r3.Vector{X: 200, Y: 100, Z: 500},
			r3.Vector{X: 150, Y: 200, Z: 0},
		),
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1000, Y: 0, Z: 0},
			r3.Vector{X: 1000, Y: 1000, Z: 0},
		),
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 1000, Y: 1000, Z: 0},
			r3.Vector{X: 0, Y: 1000, Z: 0},
		),
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1000, Z: 0},
			r3.Vector{X: 0, Y: 1000, Z: 1000},
		),
		NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 1000, Z: 1000},
			r3.Vector{X: 0, Y: 0, Z: 1000},
		),
	}

	// Create mesh with a pose and label
	originalPose := NewPose(r3.Vector{X: 100, Y: 200, Z: 300}, NewZeroOrientation())
	originalMesh := NewMesh(originalPose, triangles, "test_mesh_from_triangles")

	// Convert to protobuf
	proto := originalMesh.ToProtobuf()
	test.That(t, proto, test.ShouldNotBeNil)
	test.That(t, proto.Label, test.ShouldEqual, "test_mesh_from_triangles")

	// Restore from protobuf
	restoredGeometry, err := NewGeometryFromProto(proto)
	test.That(t, err, test.ShouldBeNil)
	restoredMesh, ok := restoredGeometry.(*Mesh)
	test.That(t, ok, test.ShouldBeTrue)

	test.That(t, restoredMesh.Label(), test.ShouldEqual, originalMesh.Label())
	test.That(t, PoseAlmostEqual(restoredMesh.Pose(), originalMesh.Pose()), test.ShouldBeTrue)
	test.That(t, len(restoredMesh.Triangles()), test.ShouldEqual, len(originalMesh.Triangles()))

	// Verify all triangles match
	originalTriangles := originalMesh.Triangles()
	restoredTriangles := restoredMesh.Triangles()
	for i, originalTri := range originalTriangles {
		restoredTri := restoredTriangles[i]
		origPoints := originalTri.Points()
		restoredPoints := restoredTri.Points()

		test.That(t, len(restoredPoints), test.ShouldEqual, len(origPoints))

		for j, origPoint := range origPoints {
			restoredPoint := restoredPoints[j]
			// The conversion from mm to meters and back can create micrometer-level float changes
			epsilon := 1e-4
			test.That(t, math.Abs(origPoint.X-restoredPoint.X), test.ShouldBeLessThan, epsilon)
			test.That(t, math.Abs(origPoint.Y-restoredPoint.Y), test.ShouldBeLessThan, epsilon)
			test.That(t, math.Abs(origPoint.Z-restoredPoint.Z), test.ShouldBeLessThan, epsilon)
		}
	}

	// Verify that the mesh can be converted to protobuf again
	secondProto := restoredMesh.ToProtobuf()
	test.That(t, secondProto, test.ShouldNotBeNil)
	test.That(t, secondProto.Label, test.ShouldEqual, originalMesh.Label())

	// Verify the protobuf content is the same
	test.That(t, secondProto.GetMesh().ContentType, test.ShouldEqual, proto.GetMesh().ContentType)
	test.That(t, len(secondProto.GetMesh().Mesh), test.ShouldEqual, len(proto.GetMesh().Mesh))
}

func TestBoxTriangleIntersectionArea(t *testing.T) {
	b, err := NewBox(NewZeroPose(), r3.Vector{X: 2, Y: 2, Z: 2}, "")
	bbox, ok := b.(*box)
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, err, test.ShouldBeNil)
	t.Run("Fully encompassed triangle", func(t *testing.T) {
		triangle := NewTriangle(
			r3.Vector{X: -0.5, Y: 0, Z: 0},
			r3.Vector{X: 0.5, Y: 0, Z: 0},
			r3.Vector{X: 0, Y: 0, Z: 0.5},
		)
		area, err := boxTriangleIntersectionArea(bbox, triangle)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, area, test.ShouldAlmostEqual, 0.25)
	})
	t.Run("Partially encompassed triangle with vertex in box", func(t *testing.T) {
		triangle := NewTriangle(
			r3.Vector{X: -1, Y: 0, Z: -2},
			r3.Vector{X: 1, Y: 0, Z: -2},
			r3.Vector{X: 0, Y: 0, Z: 0},
		)
		area, err := boxTriangleIntersectionArea(bbox, triangle)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, area, test.ShouldAlmostEqual, 0.5)
	})
	t.Run("Partially encompassed triangle with no vertices in box", func(t *testing.T) {
		triangle := NewTriangle(
			r3.Vector{X: 0, Y: 0, Z: -2},
			r3.Vector{X: 2, Y: 0, Z: -2},
			r3.Vector{X: 2, Y: 0, Z: 1},
		)
		area, err := boxTriangleIntersectionArea(bbox, triangle)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, area, test.ShouldAlmostEqual, 0.25/3)
	})
	t.Run("Triangle against box face", func(t *testing.T) {
		triangle := NewTriangle(
			r3.Vector{X: -1, Y: 1, Z: -2},
			r3.Vector{X: 1, Y: 1, Z: -2},
			r3.Vector{X: 0, Y: 1, Z: 2},
		)
		area, err := boxTriangleIntersectionArea(bbox, triangle)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, area, test.ShouldAlmostEqual, 2)
	})
	t.Run("Triangle edge against box", func(t *testing.T) {
		triangle := NewTriangle(
			r3.Vector{X: 1, Y: 1, Z: 0},
			r3.Vector{X: 1, Y: -1, Z: 0},
			r3.Vector{X: 2, Y: 0, Z: 0},
		)
		area, err := boxTriangleIntersectionArea(bbox, triangle)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, area, test.ShouldAlmostEqual, 0)
	})
	t.Run("Triangle not intersecting box", func(t *testing.T) {
		triangle := NewTriangle(
			r3.Vector{X: -1, Y: 1.1, Z: -2},
			r3.Vector{X: 1, Y: 1.1, Z: -2},
			r3.Vector{X: 0, Y: 1.1, Z: 2},
		)
		area, err := boxTriangleIntersectionArea(bbox, triangle)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, area, test.ShouldAlmostEqual, 0)
	})
}
